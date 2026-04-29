package work

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

var (
	ErrNotInitialized  = errors.New("work store is not initialized")
	ErrNotFound        = errors.New("work item not found")
	ErrAlreadyExists   = errors.New("work store file already exists")
	ErrAlreadyAccepted = errors.New("inbox item already accepted")
	ErrStoreLocked     = errors.New("work store is locked")
)

var (
	inboxIDPattern = regexp.MustCompile(`^IN-\d{4,}$`)
	workIDPattern  = regexp.MustCompile(`^W-\d{4,}$`)
	eventIDPattern = regexp.MustCompile(`^EV-\d{4,}$`)
)

type Store struct {
	root string
	now  func() time.Time
}

type configFile struct {
	Version   int       `yaml:"version"`
	CreatedAt time.Time `yaml:"created_at"`
	UpdatedAt time.Time `yaml:"updated_at"`
	Counters  counters  `yaml:"counters"`
}

type counters struct {
	LastInbox int `yaml:"last_inbox"`
	LastWork  int `yaml:"last_work"`
	LastEvent int `yaml:"last_event"`
}

type viewsFile struct {
	Version int    `yaml:"version"`
	Views   []View `yaml:"views"`
}

type idKind struct {
	prefix string
	key    string
}

var (
	inboxIDKind = idKind{prefix: "IN", key: "inbox"}
	workIDKind  = idKind{prefix: "W", key: "work"}
	eventIDKind = idKind{prefix: "EV", key: "event"}
)

const (
	mutationLockTimeout = 10 * time.Second
	lockPollInterval    = 50 * time.Millisecond
)

const defaultStoreGitignore = `.lock
.*.tmp
items/.*.tmp/
`

// New returns a Store rooted at storePath. An empty storePath uses .work.
func New(storePath string) *Store {
	if storePath == "" {
		storePath = DefaultStoreDir
	}
	return &Store{
		root: storePath,
		now:  time.Now,
	}
}

// Root returns the filesystem path backing the store.
func (s *Store) Root() string {
	return s.root
}

// Init creates the local work store layout. Existing config and view files are
// left intact, making Init safe to run repeatedly.
func (s *Store) Init() error {
	return s.withMutationLock(func() error {
		now := s.timestamp()
		for _, dir := range []string{
			s.root,
			filepath.Join(s.root, "inbox"),
			filepath.Join(s.root, "items"),
		} {
			if err := os.MkdirAll(dir, 0o755); err != nil {
				return err
			}
		}

		cfgPath := s.configPath()
		if _, err := os.Stat(cfgPath); err != nil {
			if !errors.Is(err, os.ErrNotExist) {
				return err
			}
			cfg := configFile{
				Version:   1,
				CreatedAt: now,
				UpdatedAt: now,
				Counters:  counters{},
			}
			if err := writeYAMLFile(cfgPath, cfg); err != nil {
				return err
			}
		}

		viewsPath := s.viewsPath()
		if _, err := os.Stat(viewsPath); err != nil {
			if !errors.Is(err, os.ErrNotExist) {
				return err
			}
			if err := writeYAMLFile(viewsPath, viewsFile{Version: 1, Views: defaultViews()}); err != nil {
				return err
			}
		}

		gitignorePath := s.gitignorePath()
		if _, err := os.Stat(gitignorePath); err != nil {
			if !errors.Is(err, os.ErrNotExist) {
				return err
			}
			if err := writeFileAtomic(gitignorePath, []byte(defaultStoreGitignore)); err != nil {
				return err
			}
		}

		return nil
	})
}

// AddInboxItem captures an untriaged item in .work/inbox.
func (s *Store) AddInboxItem(input InboxItemInput) (InboxItem, error) {
	title := strings.TrimSpace(input.Title)
	if title == "" {
		return InboxItem{}, errors.New("inbox item title is required")
	}
	var item InboxItem
	err := s.withMutationLock(func() error {
		if err := s.ensureInitialized(); err != nil {
			return err
		}

		id, err := s.nextID(inboxIDKind)
		if err != nil {
			return err
		}
		now := s.timestamp()
		item = InboxItem{
			ID:        id,
			Title:     title,
			Body:      strings.TrimSpace(input.Body),
			Source:    strings.TrimSpace(input.Source),
			Status:    InboxStatusOpen,
			Labels:    cloneStrings(input.Labels),
			Metadata:  cloneStringMap(input.Metadata),
			CreatedAt: now,
			UpdatedAt: now,
		}
		return writeNewYAMLFile(s.inboxPath(id), item)
	})
	return item, err
}

// ListInbox returns open inbox items sorted by ID.
func (s *Store) ListInbox() ([]InboxItem, error) {
	if err := s.ensureInitialized(); err != nil {
		return nil, err
	}
	entries, err := os.ReadDir(s.inboxDir())
	if err != nil {
		return nil, err
	}

	items := make([]InboxItem, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".yaml" {
			continue
		}
		var item InboxItem
		path := filepath.Join(s.inboxDir(), entry.Name())
		if err := readYAMLFile(path, &item); err != nil {
			return nil, fmt.Errorf("read inbox item %s: %w", path, err)
		}
		if item.Status == "" {
			item.Status = InboxStatusOpen
		}
		if item.Status == InboxStatusAccepted {
			continue
		}
		items = append(items, item)
	}
	sort.Slice(items, func(i, j int) bool { return items[i].ID < items[j].ID })
	return items, nil
}

// GetInboxItem returns one inbox item by ID.
func (s *Store) GetInboxItem(id string) (InboxItem, error) {
	if !inboxIDPattern.MatchString(id) {
		return InboxItem{}, fmt.Errorf("invalid inbox id %q", id)
	}
	if err := s.ensureInitialized(); err != nil {
		return InboxItem{}, err
	}
	return s.readInboxItem(id)
}

// AcceptInboxItem turns an inbox item into a work item and marks the inbox item
// accepted.
func (s *Store) AcceptInboxItem(id string, opts AcceptInboxOptions) (WorkItem, error) {
	if !inboxIDPattern.MatchString(id) {
		return WorkItem{}, fmt.Errorf("invalid inbox id %q", id)
	}
	var workItem WorkItem
	err := s.withMutationLock(func() error {
		if err := s.ensureInitialized(); err != nil {
			return err
		}

		inbox, err := s.readInboxItem(id)
		if err != nil {
			return err
		}
		if inbox.Status == InboxStatusAccepted {
			return fmt.Errorf("%w: %s", ErrAlreadyAccepted, id)
		}

		if existing, ok, err := s.findWorkItemBySourceInboxID(id); err != nil {
			return err
		} else if ok {
			workItem = existing
			return s.markInboxAccepted(inbox, existing.ID)
		}

		title := strings.TrimSpace(opts.Title)
		if title == "" {
			title = inbox.Title
		}
		description := strings.TrimSpace(opts.Description)
		if description == "" {
			description = inbox.Body
		}
		labels := mergeStrings(inbox.Labels, opts.Labels)
		input := WorkItemInput{
			Title:              title,
			Description:        description,
			Status:             opts.Status,
			Priority:           opts.Priority,
			Area:               opts.Area,
			Labels:             labels,
			AcceptanceCriteria: cloneStrings(opts.AcceptanceCriteria),
			Relations:          cloneRelations(opts.Relations),
			SourceInboxID:      inbox.ID,
			Metadata:           cloneStringMap(opts.Metadata),
		}
		workItem, err = s.createWorkItemLocked(input, EventInboxAccepted, map[string]any{"inbox_id": inbox.ID})
		if err != nil {
			return err
		}
		return s.markInboxAccepted(inbox, workItem.ID)
	})
	return workItem, err
}

// CreateWorkItem creates a work item directly, bypassing the inbox.
func (s *Store) CreateWorkItem(input WorkItemInput) (WorkItem, error) {
	var item WorkItem
	err := s.withMutationLock(func() error {
		if err := s.ensureInitialized(); err != nil {
			return err
		}
		var err error
		item, err = s.createWorkItemLocked(input, EventWorkCreated, nil)
		return err
	})
	return item, err
}

// ListWorkItems returns work items matching filter, sorted by ID.
func (s *Store) ListWorkItems(filter WorkItemFilter) ([]WorkItem, error) {
	if err := s.ensureInitialized(); err != nil {
		return nil, err
	}
	entries, err := os.ReadDir(s.itemsDir())
	if err != nil {
		return nil, err
	}

	items := make([]WorkItem, 0, len(entries))
	for _, entry := range entries {
		if !entry.IsDir() || !workIDPattern.MatchString(entry.Name()) {
			continue
		}
		item, err := s.readWorkItem(entry.Name())
		if err != nil {
			return nil, err
		}
		if matchesFilter(item, filter) {
			items = append(items, item)
		}
	}
	sort.Slice(items, func(i, j int) bool { return items[i].ID < items[j].ID })
	return items, nil
}

// GetWorkItem returns a work item and its append-only events.
func (s *Store) GetWorkItem(id string) (WorkItem, []Event, error) {
	if !workIDPattern.MatchString(id) {
		return WorkItem{}, nil, fmt.Errorf("invalid work item id %q", id)
	}
	if err := s.ensureInitialized(); err != nil {
		return WorkItem{}, nil, err
	}
	item, err := s.readWorkItem(id)
	if err != nil {
		return WorkItem{}, nil, err
	}
	events, err := s.readEvents(id)
	if err != nil {
		return WorkItem{}, nil, err
	}
	return item, events, nil
}

// ListView materializes the saved view identified by ID or case-insensitive
// name.
func (s *Store) ListView(idOrName string) (ViewResult, error) {
	if err := s.ensureInitialized(); err != nil {
		return ViewResult{}, err
	}
	views, err := s.readViews()
	if err != nil {
		return ViewResult{}, err
	}
	view, ok := findView(views, idOrName)
	if !ok {
		return ViewResult{}, fmt.Errorf("%w: view %q", ErrNotFound, idOrName)
	}
	items, err := s.ListWorkItems(view.Filter)
	if err != nil {
		return ViewResult{}, err
	}
	return ViewResult{View: view, Items: items}, nil
}

func (s *Store) createWorkItemLocked(input WorkItemInput, eventType EventType, eventData map[string]any) (WorkItem, error) {
	title := strings.TrimSpace(input.Title)
	if title == "" {
		return WorkItem{}, errors.New("work item title is required")
	}
	status := input.Status
	if status == "" {
		status = WorkStatusReady
	}

	id, err := s.nextID(workIDKind)
	if err != nil {
		return WorkItem{}, err
	}
	now := s.timestamp()
	item := WorkItem{
		ID:                 id,
		Title:              title,
		Description:        strings.TrimSpace(input.Description),
		Status:             status,
		Priority:           strings.TrimSpace(input.Priority),
		Area:               strings.TrimSpace(input.Area),
		Labels:             cloneStrings(input.Labels),
		AcceptanceCriteria: cloneStrings(input.AcceptanceCriteria),
		Relations:          cloneRelations(input.Relations),
		SourceInboxID:      strings.TrimSpace(input.SourceInboxID),
		Metadata:           cloneStringMap(input.Metadata),
		CreatedAt:          now,
		UpdatedAt:          now,
	}

	eventID, err := s.nextID(eventIDKind)
	if err != nil {
		return WorkItem{}, err
	}
	tempDir, err := os.MkdirTemp(s.itemsDir(), "."+id+".*.tmp")
	if err != nil {
		return WorkItem{}, err
	}
	defer func() {
		if tempDir != "" {
			_ = os.RemoveAll(tempDir)
		}
	}()

	if err := os.MkdirAll(filepath.Join(tempDir, "evidence"), 0o755); err != nil {
		return WorkItem{}, err
	}
	if err := writeYAMLFile(filepath.Join(tempDir, "item.yaml"), item); err != nil {
		return WorkItem{}, err
	}
	if err := appendEvent(filepath.Join(tempDir, "events.jsonl"), Event{
		ID:         eventID,
		WorkItemID: id,
		Type:       eventType,
		At:         now,
		Data:       cloneAnyMap(eventData),
	}); err != nil {
		return WorkItem{}, err
	}
	if _, err := os.Stat(s.itemDir(id)); err == nil {
		return WorkItem{}, fmt.Errorf("%w: %s", ErrAlreadyExists, s.itemDir(id))
	} else if !errors.Is(err, os.ErrNotExist) {
		return WorkItem{}, err
	}
	if err := os.Rename(tempDir, s.itemDir(id)); err != nil {
		return WorkItem{}, err
	}
	tempDir = ""
	return item, nil
}

func (s *Store) markInboxAccepted(inbox InboxItem, workItemID string) error {
	now := s.timestamp()
	inbox.Status = InboxStatusAccepted
	inbox.AcceptedAs = workItemID
	inbox.AcceptedAt = &now
	inbox.UpdatedAt = now
	return writeYAMLFile(s.inboxPath(inbox.ID), inbox)
}

func (s *Store) findWorkItemBySourceInboxID(inboxID string) (WorkItem, bool, error) {
	entries, err := os.ReadDir(s.itemsDir())
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return WorkItem{}, false, nil
		}
		return WorkItem{}, false, err
	}
	for _, entry := range entries {
		if !entry.IsDir() || !workIDPattern.MatchString(entry.Name()) {
			continue
		}
		item, err := s.readWorkItem(entry.Name())
		if err != nil {
			return WorkItem{}, false, err
		}
		if item.SourceInboxID == inboxID {
			return item, true, nil
		}
	}
	return WorkItem{}, false, nil
}

func (s *Store) readInboxItem(id string) (InboxItem, error) {
	var item InboxItem
	path := s.inboxPath(id)
	if err := readYAMLFile(path, &item); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return InboxItem{}, fmt.Errorf("%w: inbox item %s", ErrNotFound, id)
		}
		return InboxItem{}, err
	}
	if item.Status == "" {
		item.Status = InboxStatusOpen
	}
	return item, nil
}

func (s *Store) readWorkItem(id string) (WorkItem, error) {
	var item WorkItem
	path := s.workItemPath(id)
	if err := readYAMLFile(path, &item); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return WorkItem{}, fmt.Errorf("%w: work item %s", ErrNotFound, id)
		}
		return WorkItem{}, fmt.Errorf("read work item %s: %w", path, err)
	}
	return item, nil
}

func (s *Store) readEvents(workItemID string) ([]Event, error) {
	path := s.eventsPath(workItemID)
	file, err := os.Open(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return []Event{}, nil
		}
		return nil, err
	}
	defer file.Close()

	var events []Event
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		var event Event
		if err := json.Unmarshal([]byte(line), &event); err != nil {
			return nil, fmt.Errorf("decode event in %s: %w", path, err)
		}
		events = append(events, event)
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return events, nil
}

func (s *Store) readViews() ([]View, error) {
	var file viewsFile
	if err := readYAMLFile(s.viewsPath(), &file); err != nil {
		return nil, err
	}
	return file.Views, nil
}

func (s *Store) nextID(kind idKind) (string, error) {
	cfg, err := s.readConfig()
	if err != nil {
		return "", err
	}

	maxSeen, err := s.scanMaxID(kind)
	if err != nil {
		return "", err
	}
	last := cfg.last(kind)
	if maxSeen > last {
		last = maxSeen
	}
	next := last + 1
	cfg.setLast(kind, next)
	cfg.UpdatedAt = s.timestamp()
	if err := writeYAMLFile(s.configPath(), cfg); err != nil {
		return "", err
	}
	return fmt.Sprintf("%s-%04d", kind.prefix, next), nil
}

func (s *Store) readConfig() (configFile, error) {
	var cfg configFile
	if err := readYAMLFile(s.configPath(), &cfg); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return configFile{}, ErrNotInitialized
		}
		return configFile{}, err
	}
	if cfg.Version == 0 {
		cfg.Version = 1
	}
	return cfg, nil
}

func (s *Store) ensureInitialized() error {
	if _, err := s.readConfig(); err != nil {
		return err
	}
	for _, dir := range []string{s.inboxDir(), s.itemsDir()} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return err
		}
	}
	return nil
}

func (s *Store) scanMaxID(kind idKind) (int, error) {
	switch kind.key {
	case "inbox":
		return scanMaxFromDir(s.inboxDir(), inboxIDPattern)
	case "work":
		return scanMaxFromDir(s.itemsDir(), workIDPattern)
	case "event":
		return s.scanMaxEventID()
	default:
		return 0, fmt.Errorf("unknown id kind %q", kind.key)
	}
}

func scanMaxFromDir(dir string, pattern *regexp.Regexp) (int, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return 0, nil
		}
		return 0, err
	}
	maxID := 0
	for _, entry := range entries {
		name := strings.TrimSuffix(entry.Name(), filepath.Ext(entry.Name()))
		if !pattern.MatchString(name) {
			continue
		}
		n, err := numericID(name)
		if err != nil {
			return 0, err
		}
		if n > maxID {
			maxID = n
		}
	}
	return maxID, nil
}

func (s *Store) scanMaxEventID() (int, error) {
	entries, err := os.ReadDir(s.itemsDir())
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return 0, nil
		}
		return 0, err
	}
	maxID := 0
	for _, entry := range entries {
		if !entry.IsDir() || !workIDPattern.MatchString(entry.Name()) {
			continue
		}
		events, err := s.readEvents(entry.Name())
		if err != nil {
			return 0, err
		}
		for _, event := range events {
			if !eventIDPattern.MatchString(event.ID) {
				continue
			}
			n, err := numericID(event.ID)
			if err != nil {
				return 0, err
			}
			if n > maxID {
				maxID = n
			}
		}
	}
	return maxID, nil
}

func (cfg configFile) last(kind idKind) int {
	switch kind.key {
	case "inbox":
		return cfg.Counters.LastInbox
	case "work":
		return cfg.Counters.LastWork
	case "event":
		return cfg.Counters.LastEvent
	default:
		return 0
	}
}

func (cfg *configFile) setLast(kind idKind, n int) {
	switch kind.key {
	case "inbox":
		cfg.Counters.LastInbox = n
	case "work":
		cfg.Counters.LastWork = n
	case "event":
		cfg.Counters.LastEvent = n
	}
}

func numericID(id string) (int, error) {
	_, raw, ok := strings.Cut(id, "-")
	if !ok {
		return 0, fmt.Errorf("invalid id %q", id)
	}
	return strconv.Atoi(raw)
}

func matchesFilter(item WorkItem, filter WorkItemFilter) bool {
	if len(filter.IDs) > 0 && !containsString(filter.IDs, item.ID) {
		return false
	}
	if len(filter.Statuses) > 0 && !containsStatus(filter.Statuses, item.Status) {
		return false
	}
	if len(filter.Areas) > 0 && !containsString(filter.Areas, item.Area) {
		return false
	}
	if len(filter.Labels) > 0 && !containsAllStrings(item.Labels, filter.Labels) {
		return false
	}
	if strings.TrimSpace(filter.Text) != "" {
		haystack := strings.ToLower(item.Title + "\n" + item.Description)
		if !strings.Contains(haystack, strings.ToLower(strings.TrimSpace(filter.Text))) {
			return false
		}
	}
	return true
}

func findView(views []View, idOrName string) (View, bool) {
	needle := strings.TrimSpace(idOrName)
	for _, view := range views {
		if view.ID == needle || strings.EqualFold(view.Name, needle) {
			return view, true
		}
	}
	return View{}, false
}

func defaultViews() []View {
	return []View{
		{
			ID:          "all",
			Name:        "All",
			Description: "All work items",
		},
		{
			ID:          "ready",
			Name:        "Ready",
			Description: "Ready work that can be claimed",
			Filter: WorkItemFilter{
				Statuses: []WorkStatus{WorkStatusReady},
			},
		},
		{
			ID:          "active",
			Name:        "Active",
			Description: "Active work currently in motion",
			Filter: WorkItemFilter{
				Statuses: []WorkStatus{WorkStatusActive},
			},
		},
		{
			ID:          "blocked",
			Name:        "Blocked",
			Description: "Work waiting on an unblock condition",
			Filter: WorkItemFilter{
				Statuses: []WorkStatus{WorkStatusBlocked},
			},
		},
		{
			ID:          "done",
			Name:        "Done",
			Description: "Completed work",
			Filter: WorkItemFilter{
				Statuses: []WorkStatus{WorkStatusDone},
			},
		},
	}
}

func appendEvent(path string, event Event) error {
	if event.ID == "" {
		return errors.New("event id is required")
	}
	if event.WorkItemID == "" {
		return errors.New("event work item id is required")
	}
	if event.Type == "" {
		return errors.New("event type is required")
	}
	if event.At.IsZero() {
		return errors.New("event timestamp is required")
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	data, err := json.Marshal(event)
	if err != nil {
		return err
	}
	file, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	defer file.Close()
	if _, err := file.Write(append(data, '\n')); err != nil {
		return err
	}
	return nil
}

type storeLock struct {
	file *os.File
	path string
}

func (s *Store) withMutationLock(fn func() error) (err error) {
	if err := os.MkdirAll(s.root, 0o755); err != nil {
		return err
	}
	lock, err := acquireStoreLock(s.lockPath(), mutationLockTimeout)
	if err != nil {
		return err
	}
	defer func() {
		if unlockErr := lock.release(); err == nil {
			err = unlockErr
		}
	}()
	return fn()
}

func acquireStoreLock(path string, timeout time.Duration) (*storeLock, error) {
	deadline := time.Now().Add(timeout)
	for {
		file, err := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o644)
		if err == nil {
			_, _ = fmt.Fprintf(file, "pid: %d\ncreated_at: %s\n", os.Getpid(), time.Now().UTC().Format(time.RFC3339Nano))
			return &storeLock{file: file, path: path}, nil
		}
		if !errors.Is(err, os.ErrExist) {
			return nil, err
		}
		if time.Now().After(deadline) {
			return nil, fmt.Errorf("%w: %s", ErrStoreLocked, path)
		}
		time.Sleep(lockPollInterval)
	}
}

func (l *storeLock) release() error {
	closeErr := l.file.Close()
	removeErr := os.Remove(l.path)
	if errors.Is(removeErr, os.ErrNotExist) {
		removeErr = nil
	}
	return errors.Join(closeErr, removeErr)
}

func readYAMLFile(path string, out any) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return yaml.Unmarshal(data, out)
}

func writeNewYAMLFile(path string, value any) error {
	if _, err := os.Stat(path); err == nil {
		return fmt.Errorf("%w: %s", ErrAlreadyExists, path)
	} else if !errors.Is(err, os.ErrNotExist) {
		return err
	}
	return writeYAMLFile(path, value)
}

func writeYAMLFile(path string, value any) error {
	data, err := yaml.Marshal(value)
	if err != nil {
		return err
	}
	return writeFileAtomic(path, data)
}

func writeFileAtomic(path string, data []byte) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	temp, err := os.CreateTemp(dir, "."+filepath.Base(path)+".*.tmp")
	if err != nil {
		return err
	}
	tempName := temp.Name()
	defer os.Remove(tempName)

	if _, err := temp.Write(data); err != nil {
		temp.Close()
		return err
	}
	if err := temp.Chmod(0o644); err != nil {
		temp.Close()
		return err
	}
	if err := temp.Close(); err != nil {
		return err
	}
	return os.Rename(tempName, path)
}

func (s *Store) timestamp() time.Time {
	return s.now().UTC().Round(0)
}

func (s *Store) configPath() string {
	return filepath.Join(s.root, "config.yaml")
}

func (s *Store) viewsPath() string {
	return filepath.Join(s.root, "views.yaml")
}

func (s *Store) gitignorePath() string {
	return filepath.Join(s.root, ".gitignore")
}

func (s *Store) inboxDir() string {
	return filepath.Join(s.root, "inbox")
}

func (s *Store) itemsDir() string {
	return filepath.Join(s.root, "items")
}

func (s *Store) lockPath() string {
	return filepath.Join(s.root, ".lock")
}

func (s *Store) inboxPath(id string) string {
	return filepath.Join(s.inboxDir(), id+".yaml")
}

func (s *Store) itemDir(id string) string {
	return filepath.Join(s.itemsDir(), id)
}

func (s *Store) workItemPath(id string) string {
	return filepath.Join(s.itemDir(id), "item.yaml")
}

func (s *Store) eventsPath(id string) string {
	return filepath.Join(s.itemDir(id), "events.jsonl")
}

func (s *Store) evidenceDir(id string) string {
	return filepath.Join(s.itemDir(id), "evidence")
}

func cloneStrings(in []string) []string {
	if len(in) == 0 {
		return nil
	}
	out := append([]string(nil), in...)
	return out
}

func cloneStringMap(in map[string]string) map[string]string {
	if len(in) == 0 {
		return nil
	}
	out := make(map[string]string, len(in))
	for k, v := range in {
		out[k] = v
	}
	return out
}

func cloneAnyMap(in map[string]any) map[string]any {
	if len(in) == 0 {
		return nil
	}
	out := make(map[string]any, len(in))
	for k, v := range in {
		out[k] = v
	}
	return out
}

func cloneRelations(in []Relation) []Relation {
	if len(in) == 0 {
		return nil
	}
	out := append([]Relation(nil), in...)
	return out
}

func mergeStrings(a, b []string) []string {
	if len(a) == 0 {
		return cloneStrings(b)
	}
	if len(b) == 0 {
		return cloneStrings(a)
	}
	seen := make(map[string]bool, len(a)+len(b))
	out := make([]string, 0, len(a)+len(b))
	for _, value := range append(cloneStrings(a), b...) {
		if seen[value] {
			continue
		}
		seen[value] = true
		out = append(out, value)
	}
	return out
}

func containsString(items []string, target string) bool {
	for _, item := range items {
		if item == target {
			return true
		}
	}
	return false
}

func containsStatus(items []WorkStatus, target WorkStatus) bool {
	for _, item := range items {
		if item == target {
			return true
		}
	}
	return false
}

func containsAllStrings(haystack, needles []string) bool {
	for _, needle := range needles {
		if !containsString(haystack, needle) {
			return false
		}
	}
	return true
}
