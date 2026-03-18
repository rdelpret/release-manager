package template

// TemplateList defines a task list with its groups and tasks
type TemplateList struct {
	Name   string
	Color  string
	Groups []TemplateGroup
}

type TemplateGroup struct {
	Name  string
	Tasks []TemplateTask
}

type TemplateTask struct {
	Name     string
	Subtasks []string
}

// DefaultTemplate returns the 5 pre-loaded task lists for a new campaign
func DefaultTemplate() []TemplateList {
	return []TemplateList{
		{
			Name:  "Campaign Assets",
			Color: "#3B82F6",
			Groups: []TemplateGroup{
				{
					Name: "Song Assets",
					Tasks: []TemplateTask{
						{Name: "Make New Asset Folder"},
						{Name: "Release Date"},
						{Name: "Final Master"},
						{Name: "Final Artwork"},
						{Name: "Instrumental"},
						{Name: "Stems"},
						{Name: "Lyrics"},
						{Name: "Press Photos"},
						{Name: "Music Video/Lyric Video"},
						{Name: "Folder for Promo Content"},
					},
				},
			},
		},
		{
			Name:  "Campaign Tasklist",
			Color: "#22C55E",
			Groups: []TemplateGroup{
				{
					Name: "Pre-Release",
					Tasks: []TemplateTask{
						{Name: "Campaign Setup + Timeline"},
						{Name: "Contracts"},
						{Name: "Photoshoot for release"},
						{Name: "Make content plan"},
						{Name: "Create content or work with a graphic designer"},
						{Name: "Create pre-save Link"},
						{Name: "Pre-save campaign/contest"},
						{Name: "Create linktree"},
						{Name: "Write out hashtags"},
						{Name: "Update website - coming soon"},
						{Name: "Update all SM profiles - coming soon"},
						{Name: "Update streaming profiles"},
					},
				},
			},
		},
		{
			Name:  "Social Media Content",
			Color: "#A78BFA",
			Groups: []TemplateGroup{
				{
					Name: "Pre-Release Content",
					Tasks: []TemplateTask{
						{Name: "Teaser post"},
						{Name: "Banner - release date - presave now"},
						{Name: "Single art - story sized - release date or 'pre-save now'"},
						{Name: "Single art edit #1 (with release date)"},
						{Name: "Pre-save Post"},
						{Name: "Song poster #1 - storytelling"},
						{Name: "Song teaser #1"},
						{Name: "\"Midnight\" or \"Tomorrow\""},
					},
				},
				{
					Name: "Post-Release Content",
					Tasks: []TemplateTask{
						{Name: "Banner - out now"},
						{Name: "Song Clip - out now"},
						{Name: "Single art edit #2 (\"out now\")"},
						{Name: "Single art - out now - story sized"},
						{Name: "Song poster #2"},
						{Name: "Studio content #1 - storytelling"},
						{Name: "Single art edit #3 - storytelling"},
						{Name: "Credits page"},
						{Name: "Lyric poster/Lyric Reel"},
					},
				},
				{
					Name: "Extra Content",
					Tasks: []TemplateTask{
						{Name: "Pre-save Post #2"},
						{Name: "Song clip/teaser #3"},
						{Name: "Song poster #3"},
						{Name: "Studio content #2"},
						{Name: "Tiktok/Reels"},
						{Name: "Clips from Music video/lyric video"},
						{Name: "Countdowns"},
						{Name: "Acoustic version/demo version/live version"},
					},
				},
			},
		},
		{
			Name:  "PR",
			Color: "#EAB308",
			Groups: []TemplateGroup{
				{
					Name: "Media Outreach",
					Tasks: []TemplateTask{
						{Name: "Research new outlets & develop CRM"},
						{Name: "Create press release"},
						{Name: "Create e-mail outreach pitch"},
						{Name: "Traditional Media outreach"},
						{Name: "New Media outreach"},
						{Name: "Submit to third party platforms"},
						{Name: "Press follow ups"},
					},
				},
				{
					Name: "Playlisting",
					Tasks: []TemplateTask{
						{Name: "Research playlists"},
						{Name: "Write playlist pitch"},
						{Name: "Pitch to independent curators"},
						{Name: "Submit to third party platforms"},
					},
				},
				{
					Name: "Community Engagement",
					Tasks: []TemplateTask{
						{Name: "Send DMs on IG/Twitter/FB with link"},
						{Name: "Post on discord chats/reddit threads"},
						{Name: "Send mailing list e-mail - presave"},
						{Name: "Send mailing list e-mail - out now"},
					},
				},
			},
		},
		{
			Name:  "Distribution",
			Color: "#EF4444",
			Groups: []TemplateGroup{
				{
					Name: "To Do",
					Tasks: []TemplateTask{
						{Name: "Obtain license for cover song"},
						{Name: "Upload to distributor", Subtasks: []string{"Fill out credits", "Upload lyrics"}},
						{Name: "Pitch song to Spotify editors"},
						{Name: "Upload Spotify Canvas"},
						{Name: "Register song with PROs"},
						{Name: "For physical releases"},
					},
				},
			},
		},
	}
}
