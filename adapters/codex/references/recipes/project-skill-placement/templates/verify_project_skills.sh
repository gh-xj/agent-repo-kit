#!/usr/bin/env bash
# Validate project-local skill discovery wiring and policy-backed allowlist.
set -euo pipefail

cd "$(git rev-parse --show-toplevel)"

POLICY_PATH="${PROJECT_SKILL_POLICY:-tools/skills/project-skill-policy.yml}"

failed=0
warned=0

fail() {
  echo "FAIL: $*" >&2
  failed=1
}

warn() {
  echo "WARN: $*" >&2
  warned=1
}

check_no_untracked_skill_entries() {
  local canonical_root mirror_root untracked

  canonical_root="$(
    ruby -ryaml -e 'policy = YAML.safe_load(File.read(ARGV.fetch(0)), permitted_classes: [], aliases: false); puts policy["canonicalRoot"].to_s' "$POLICY_PATH"
  )"
  mirror_root="$(
    ruby -ryaml -e 'policy = YAML.safe_load(File.read(ARGV.fetch(0)), permitted_classes: [], aliases: false); puts policy["mirrorRoot"].to_s' "$POLICY_PATH"
  )"

  untracked="$(git ls-files --others --exclude-standard "$canonical_root" "$mirror_root")"
  [ -z "$untracked" ] && return 0

  fail "untracked project skill entries found; either remove them or declare and track them:"
  printf '%s\n' "$untracked" >&2
}

check_policy() {
  [ -f "$POLICY_PATH" ] || {
    fail "$POLICY_PATH is missing"
    return 0
  }

  local output
  if ! output="$(
    ruby -ryaml -rjson -rdigest -rpathname -e '
      policy_path = ARGV.fetch(0)

      def emit(kind, message)
        puts [kind, message].join("\t")
      end

      begin
        policy = YAML.safe_load(File.read(policy_path), permitted_classes: [], aliases: false)
      rescue StandardError => e
        emit("FAIL", "#{policy_path} is not valid YAML: #{e.message}")
        exit
      end

      unless policy.is_a?(Hash)
        emit("FAIL", "#{policy_path} must contain a YAML mapping")
        exit
      end

      emit("FAIL", "#{policy_path} version must be 1") unless policy["version"] == 1

      canonical_root = policy["canonicalRoot"].to_s
      mirror_root = policy["mirrorRoot"].to_s
      mirror_rule = policy["mirrorRule"].to_s

      emit("FAIL", "#{policy_path} canonicalRoot is missing") if canonical_root.empty?
      emit("FAIL", "#{policy_path} mirrorRoot is missing") if mirror_root.empty?
      emit("FAIL", "#{policy_path} mirrorRule must be symlink-to-canonical") unless mirror_rule == "symlink-to-canonical"

      skills = policy["skills"]
      unless skills.is_a?(Hash) && !skills.empty?
        emit("FAIL", "#{policy_path} must declare a non-empty skills mapping")
        exit
      end

      [canonical_root, mirror_root].each do |root|
        emit("FAIL", "#{root} is missing") unless Dir.exist?(root)
      end

      policy_names = skills.keys.map(&:to_s).sort
      actual_skills = Dir.exist?(canonical_root) ? Dir.children(canonical_root).select { |name| File.directory?(File.join(canonical_root, name)) }.sort : []
      actual_mirrors = Dir.exist?(mirror_root) ? Dir.children(mirror_root).reject { |name| name == "README.md" }.sort : []

      (actual_skills - policy_names).each { |name| emit("FAIL", "#{canonical_root}/#{name} is not declared in #{policy_path}") }
      (actual_mirrors - policy_names).each { |name| emit("FAIL", "#{mirror_root}/#{name} is not declared in #{policy_path}") }
      (policy_names - actual_skills).each { |name| emit("FAIL", "#{policy_path} declares #{name}, but #{canonical_root}/#{name} is missing") }

      lock_data = {}
      if File.file?("skills-lock.json")
        begin
          lock_data = JSON.parse(File.read("skills-lock.json"))
        rescue StandardError => e
          emit("FAIL", "skills-lock.json is not valid JSON: #{e.message}")
        end
      end

      skills.sort.each do |name, cfg|
        cfg = {} unless cfg.is_a?(Hash)
        skill_root = File.join(canonical_root, name)
        skill_file = File.join(skill_root, "SKILL.md")
        mirror = File.join(mirror_root, name)
        expected_target = Pathname
          .new(File.join(canonical_root, name))
          .relative_path_from(Pathname.new(File.dirname(mirror)))
          .to_s

        unless File.file?(skill_file)
          emit("FAIL", "#{skill_file} is missing")
          next
        end

        text = File.read(skill_file)
        frontmatter = text.match(/\A---\n(.*?)\n---\n/m)
        if frontmatter
          data = YAML.safe_load(frontmatter[1], permitted_classes: [], aliases: false) || {}
          emit("FAIL", "#{skill_file} frontmatter name must be #{name}") unless data["name"].to_s == name
          description = data["description"].to_s.strip
          emit("WARN", "#{skill_file} description is short; trigger quality may be weak") if description.length < 30
        else
          emit("FAIL", "#{skill_file} is missing YAML frontmatter")
        end

        placement = cfg["placement"].to_s
        unless %w[wrapper-local project-exception].include?(placement)
          emit("FAIL", "#{policy_path} #{name}.placement must be wrapper-local or project-exception")
        end

        source_kind = cfg.fetch("sourceKind", "local").to_s
        unless %w[local external].include?(source_kind)
          emit("FAIL", "#{policy_path} #{name}.sourceKind must be local or external")
        end

        reason = cfg["reason"].to_s.strip
        emit("WARN", "#{policy_path} #{name}.reason is short; explain why this belongs at project level") if reason.length < 30

        if !File.symlink?(mirror)
          emit("FAIL", "#{mirror} must be a symlink to #{expected_target}")
        elsif File.readlink(mirror) != expected_target
          emit("FAIL", "#{mirror} points to #{File.readlink(mirror)}, expected #{expected_target}")
        end

        if source_kind == "external"
          upstream = cfg["upstream"].is_a?(Hash) ? cfg["upstream"] : {}
          if !File.file?("skills-lock.json")
            emit("FAIL", "#{name} is external but skills-lock.json is missing")
          else
            entry = lock_data.fetch("skills", {}).fetch(name, nil)
            if entry.nil?
              emit("FAIL", "#{name} is external but missing from skills-lock.json")
            else
              {
                "source" => upstream["source"].to_s,
                "sourceType" => upstream["sourceType"].to_s,
                "skillPath" => upstream["skillPath"].to_s
              }.each do |key, expected|
                actual = entry[key].to_s
                emit("FAIL", "#{name} skills-lock.json #{key}=#{actual.inspect}, expected #{expected.inspect}") unless actual == expected
              end

              skill_path = entry["skillPath"].to_s
              local_path = File.join(canonical_root, name, skill_path)
              if File.file?(local_path)
                actual_hash = Digest::SHA256.file(local_path).hexdigest
                expected_hash = entry["computedHash"].to_s
                emit("FAIL", "#{name} computedHash does not match #{local_path}") unless actual_hash == expected_hash
              end
            end
          end
        end

        retained_paths = Array(cfg["retainedPaths"]).map(&:to_s)
        unless retained_paths.empty?
          all_paths = Dir.glob(File.join(skill_root, "**", "*"), File::FNM_DOTMATCH)
            .reject { |path| [".", ".."].include?(File.basename(path)) }
            .map { |path| path.delete_prefix(skill_root + "/") }
          extras = all_paths.reject do |relative_path|
            retained_paths.any? do |retained_path|
              relative_path == retained_path ||
                (retained_path.end_with?("/") && relative_path.start_with?(retained_path))
            end
          end
          extras.each { |relative_path| emit("FAIL", "#{name} retains only #{retained_paths.join(", ")}; remove #{skill_root}/#{relative_path}") }
        end

        Array(cfg["cliPackages"]).each do |package_cfg|
          package_cfg = {} unless package_cfg.is_a?(Hash)
          manager = package_cfg["manager"].to_s
          package_name = package_cfg["name"].to_s
          if manager == "npm-global" && !package_name.empty?
            emit("NPM_GLOBAL", [name, package_name].join("\t"))
          else
            emit("FAIL", "#{policy_path} #{name}.cliPackages has unsupported or incomplete package entry")
          end
        end
      end
    ' "$POLICY_PATH"
  )"; then
    fail "$POLICY_PATH policy check crashed"
    printf '%s\n' "$output" >&2
    return 0
  fi

  [ -n "$output" ] || return 0

  while IFS=$'\t' read -r kind first second; do
    case "$kind" in
      FAIL) fail "$first" ;;
      WARN) warn "$first" ;;
      NPM_GLOBAL) check_npm_global_package "$first" "$second" ;;
      *) fail "$POLICY_PATH emitted unknown check record: $kind $first $second" ;;
    esac
  done <<< "$output"
}

check_npm_global_package() {
  local skill_name=$1
  local package_name=$2

  if ! command -v npm >/dev/null 2>&1; then
    fail "npm is required to verify $skill_name CLI package $package_name"
    return 0
  fi

  local npm_json
  npm_json="$(npm list -g "$package_name" --json --depth=0 2>/dev/null || true)"
  if ! printf "%s" "$npm_json" | ruby -rjson -e '
    package_name = ARGV.fetch(0)
    data = JSON.parse(STDIN.read)
    dep = data.fetch("dependencies", {}).fetch(package_name, nil)
    exit(dep && dep["version"].to_s != "" ? 0 : 1)
  ' "$package_name"; then
    fail "$skill_name project skill requires global npm package $package_name"
  fi
}

check_conventions_skill_policy() {
  [ -f .conventions.yaml ] || return 0
  [ -f "$POLICY_PATH" ] || return 0

  local output
  if ! output="$(
    ruby -ryaml -e '
      policy_path = ARGV.fetch(0)
      policy = YAML.safe_load(File.read(policy_path), permitted_classes: [], aliases: false)
      conventions = YAML.safe_load(File.read(".conventions.yaml"), permitted_classes: [], aliases: false)
      skill_policy = conventions.fetch("skill_policy", {})

      failures = []
      failures << ".conventions.yaml skill_policy.policy_file must be #{policy_path}" unless skill_policy["policy_file"].to_s == policy_path
      failures << ".conventions.yaml canonical root does not match policy" unless skill_policy["canonical_project_root"].to_s == policy["canonicalRoot"].to_s
      failures << ".conventions.yaml mirror root does not match policy" unless skill_policy["mirror_project_root"].to_s == policy["mirrorRoot"].to_s

      exception_names = Array(skill_policy["explicit_project_exceptions"]).map { |entry| entry.is_a?(Hash) ? entry["name"].to_s : nil }.compact
      policy.fetch("skills", {}).each do |name, cfg|
        next unless cfg["placement"].to_s == "project-exception"
        failures << "#{name} is a project-exception but missing from .conventions.yaml" unless exception_names.include?(name)
      end

      failures.each { |message| puts message }
    ' "$POLICY_PATH"
  )"; then
    fail ".conventions.yaml skill policy check crashed"
    return 0
  fi

  while IFS= read -r message; do
    [ -n "$message" ] || continue
    fail "$message"
  done <<< "$output"
}

check_no_untracked_skill_entries
check_policy
check_conventions_skill_policy

if [ "$failed" -ne 0 ]; then
  exit 1
fi

if [ "$warned" -ne 0 ]; then
  echo "OK: project-local skill wiring is valid (with warnings)"
else
  echo "OK: project-local skill wiring is valid"
fi
