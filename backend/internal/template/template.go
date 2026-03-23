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
					Name: "Track Assets",
					Tasks: []TemplateTask{
						{Name: "Create asset folder"},
						{Name: "Set release date"},
						{Name: "Final master (WAV + MP3)"},
						{Name: "Final artwork (3000x3000)"},
						{Name: "Extended mix"},
						{Name: "Stems for remixers"},
						{Name: "DJ-friendly intro/outro edit"},
						{Name: "Visualizer / lyric video"},
						{Name: "Music video"},
						{Name: "Promo content folder (clips, BTS, etc.)"},
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
						{Name: "Campaign setup + timeline"},
						{Name: "Contracts / splits / publishing"},
						{Name: "Create content plan"},
						{Name: "Commission artwork / visual assets"},
						{Name: "Create pre-save link"},
						{Name: "Pre-save campaign / contest"},
						{Name: "Create or update Linktree"},
						{Name: "Write out hashtags"},
						{Name: "Update website - coming soon"},
						{Name: "Update social media bios - coming soon"},
						{Name: "Update Spotify / Apple Music artist profiles"},
						{Name: "Send track to DJ promo pool"},
						{Name: "Send to radio promo contacts"},
						{Name: "Submit to Spotify editorial playlists"},
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
						{Name: "Teaser clip (15-30s preview)"},
						{Name: "Artwork reveal post"},
						{Name: "Banner - release date + pre-save"},
						{Name: "Story-sized artwork (with release date)"},
						{Name: "Pre-save post"},
						{Name: "Studio session / production breakdown clip"},
						{Name: "Behind the track post (inspiration, story)"},
						{Name: "\"Midnight\" / \"Tomorrow\" countdown post"},
					},
				},
				{
					Name: "Post-Release Content",
					Tasks: []TemplateTask{
						{Name: "Banner - out now"},
						{Name: "Track clip - out now"},
						{Name: "Artwork edit - out now"},
						{Name: "Story-sized out now graphic"},
						{Name: "Visualizer clip for socials"},
						{Name: "DJ set clip featuring the track"},
						{Name: "Studio breakdown / production tips reel"},
						{Name: "Credits / collab shoutout post"},
						{Name: "Fan reaction / UGC repost"},
					},
				},
				{
					Name: "Extra Content",
					Tasks: []TemplateTask{
						{Name: "TikTok / Reels with track"},
						{Name: "DJ mix snippet featuring the track"},
						{Name: "Live set / festival clip with track"},
						{Name: "Remix teasers (if applicable)"},
						{Name: "Countdown series"},
						{Name: "Playlist placement screenshot posts"},
						{Name: "Milestone celebration posts (streams, charts)"},
						{Name: "VIP / remix / bootleg version tease"},
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
						{Name: "Research EDM blogs & outlets"},
						{Name: "Create press release"},
						{Name: "Create email pitch"},
						{Name: "Pitch to EDM blogs (EDM.com, Your EDM, Dancing Astronaut, etc.)"},
						{Name: "Pitch to YouTube channels (MrSuicideSheep, Trap Nation, etc.)"},
						{Name: "Submit to DJ mix shows / radio (Diplo's Revolution, BBC Radio 1, etc.)"},
						{Name: "Press follow ups"},
					},
				},
				{
					Name: "Playlisting",
					Tasks: []TemplateTask{
						{Name: "Research genre-specific Spotify playlists"},
						{Name: "Write playlist pitch"},
						{Name: "Pitch to independent curators"},
						{Name: "Submit via SubmitHub / PlaylistPush / DailyPlaylists"},
						{Name: "Pitch to Apple Music / Amazon Music editorial"},
					},
				},
				{
					Name: "Community & DJ Network",
					Tasks: []TemplateTask{
						{Name: "Send to DJ friends for support"},
						{Name: "Post in producer Discord servers / Facebook groups"},
						{Name: "Post on Reddit (r/EDM, genre subs)"},
						{Name: "Send mailing list email - pre-save"},
						{Name: "Send mailing list email - out now"},
						{Name: "Engage in comments on related posts"},
					},
				},
			},
		},
		{
			Name:  "Distribution",
			Color: "#EF4444",
			Groups: []TemplateGroup{
				{
					Name: "Distribution Tasks",
					Tasks: []TemplateTask{
						{Name: "Upload to distributor", Subtasks: []string{"Fill out credits & metadata", "Set release date", "Upload artwork", "Select stores & territories"}},
						{Name: "Submit to Beatport (if applicable)"},
						{Name: "Pitch to Spotify editorial"},
						{Name: "Upload Spotify Canvas"},
						{Name: "Register with PROs (ASCAP/BMI/SESAC)"},
						{Name: "Register with SoundExchange"},
						{Name: "Set up DJ promo pool distribution"},
						{Name: "Upload remix package / stems (if applicable)"},
						{Name: "Obtain sample clearance (if applicable)"},
					},
				},
			},
		},
	}
}
