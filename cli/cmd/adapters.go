package cmd

// AdaptersCmd groups the `ark adapters` subcommands.
type AdaptersCmd struct {
	Link      AdaptersLinkCmd      `cmd:"" help:"symlink every skill under skills/ into a harness's skill root"`
	ListLinks AdaptersListLinksCmd `cmd:"list-links" help:"print <srcAbs>\\t<dstAbs> per discovered skill for a harness"`
}
