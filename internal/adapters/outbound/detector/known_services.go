package detector

// ServiceMetadata stores metadata mapping for known services.
type ServiceMetadata struct {
	Name      string
	DeleteURL string
}

// KnownServices maps domain strings to popular service names and deletion URLs.
var KnownServices = map[string]ServiceMetadata{
	"github.com":         {Name: "GitHub", DeleteURL: "https://github.com/settings/security"},
	"github.io":          {Name: "GitHub Pages", DeleteURL: "https://github.com/settings/security"},
	"google.com":         {Name: "Google", DeleteURL: "https://myaccount.google.com/delete-services-or-account"},
	"gmail.com":          {Name: "Gmail", DeleteURL: "https://myaccount.google.com/delete-services-or-account"},
	"googlemail.com":     {Name: "Google Mail", DeleteURL: "https://myaccount.google.com/delete-services-or-account"},
	"youtube.com":        {Name: "YouTube", DeleteURL: "https://myaccount.google.com/delete-services-or-account"},
	"netflix.com":        {Name: "Netflix", DeleteURL: "https://www.netflix.com/CancelPlan"},
	"spotify.com":        {Name: "Spotify", DeleteURL: "https://www.spotify.com/about-us/contact/close-account/"},
	"microsoft.com":      {Name: "Microsoft", DeleteURL: "https://account.live.com/closeaccount.aspx"},
	"outlook.com":        {Name: "Outlook / Hotmail", DeleteURL: "https://account.live.com/closeaccount.aspx"},
	"live.com":           {Name: "Microsoft Live", DeleteURL: "https://account.live.com/closeaccount.aspx"},
	"facebook.com":       {Name: "Facebook", DeleteURL: "https://www.facebook.com/deactivate_delete_account"},
	"facebookmail.com":   {Name: "Facebook Mail", DeleteURL: "https://www.facebook.com/deactivate_delete_account"},
	"instagram.com":      {Name: "Instagram", DeleteURL: "https://www.instagram.com/accounts/remove/request/permanent/"},
	"twitter.com":        {Name: "Twitter / X", DeleteURL: "https://twitter.com/settings/deactivate"},
	"x.com":              {Name: "Twitter / X", DeleteURL: "https://x.com/settings/deactivate"},
	"linkedin.com":       {Name: "LinkedIn", DeleteURL: "https://www.linkedin.com/help/linkedin/answer/a1337996"},
	"discord.com":        {Name: "Discord", DeleteURL: "https://support.discord.com/hc/en-us/articles/212500618-How-do-I-delete-my-account-"},
	"discordapp.com":     {Name: "Discord App", DeleteURL: "https://support.discord.com/hc/en-us/articles/212500618-How-do-I-delete-my-account-"},
	"zoom.us":            {Name: "Zoom", DeleteURL: "https://zoom.us/profile"},
	"slack.com":          {Name: "Slack", DeleteURL: "https://slack.com/help/articles/201905068-Deactivate-your-Slack-account"},
	"patreon.com":        {Name: "Patreon", DeleteURL: "https://support.patreon.com/hc/en-us/articles/360004153091-How-do-I-close-my-account-"},
	"steamcommunity.com":  {Name: "Steam Community", DeleteURL: "https://help.steampowered.com/en/wizard/HelpWithAccountDelete"},
	"steampowered.com":    {Name: "Steam Games", DeleteURL: "https://help.steampowered.com/en/wizard/HelpWithAccountDelete"},
	"gitlab.com":         {Name: "GitLab", DeleteURL: "https://gitlab.com/-/profile/account"},
	"trello.com":         {Name: "Trello", DeleteURL: "https://trello.com/me"},
	"adobe.com":          {Name: "Adobe", DeleteURL: "https://account.adobe.com/privacy"},
	"medium.com":         {Name: "Medium", DeleteURL: "https://medium.com/me/settings"},
	"canva.com":          {Name: "Canva", DeleteURL: "https://www.canva.com/settings/your-account"},
	"dropbox.com":        {Name: "Dropbox", DeleteURL: "https://www.dropbox.com/account/delete"},
	"reddit.com":         {Name: "Reddit", DeleteURL: "https://www.reddit.com/settings/data-deletion"},
	"pinterest.com":      {Name: "Pinterest", DeleteURL: "https://www.pinterest.com/settings/deactivate"},
	"tumblr.com":         {Name: "Tumblr", DeleteURL: "https://www.tumblr.com/settings/account"},
	"quora.com":          {Name: "Quora", DeleteURL: "https://www.quora.com/settings/privacy"},
	"vimeo.com":          {Name: "Vimeo", DeleteURL: "https://vimeo.com/settings/membership"},
	"figma.com":          {Name: "Figma", DeleteURL: "https://www.figma.com/settings"},
	"notion.so":          {Name: "Notion", DeleteURL: "https://www.notion.so/my-settings"},
	"wikimedia.org":      {Name: "Wikimedia / Wikipedia", DeleteURL: "https://meta.wikimedia.org/wiki/Right_to_be_forgotten"},
	"wikipedia.org":      {Name: "Wikipedia", DeleteURL: "https://meta.wikimedia.org/wiki/Right_to_be_forgotten"},
}
