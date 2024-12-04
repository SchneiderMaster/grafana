package kind

name:        "Preferences"
pluralName:  "Preferences"
maturity:    "merged"
description: "The user or team frontend preferences"

lineage: schemas: [{
	version: [0, 0]
	schema: {
		// Spec defines user, team or org Grafana preferences
		// swagger:model Preferences
		spec: {
			// UID for the home dashboard
			homeDashboardUID?: string

			// The timezone selection
			// TODO: this should use the timezone defined in common
			timezone?: string

			// day of the week (sunday, monday, etc)
			weekStart?: string

			// light, dark, empty is default
			theme?: string

			// Selected language (beta)
			language?: string

			// Explore query history preferences
			queryHistory?: #QueryHistoryPreference

			// Cookie preferences
			cookiePreferences?: #CookiePreferences

			// Navigation preferences
			navbar?: #NavbarPreference

			// Custom commands for commands palette
			customCommands?: [...#CustomCommand]
		} @cuetsy(kind="interface")

		#QueryHistoryPreference: {
			// one of: '' | 'query' | 'starred';
			homeTab?: string
		} @cuetsy(kind="interface")

		#CookiePreferences: {
			analytics?: {}
			performance?: {}
			functional?: {}
		} @cuetsy(kind="interface")

		#NavbarPreference: {
			bookmarkUrls: [...string]
		} @cuetsy(kind="interface")

		#CustomCommand: {
			// Unique identifier for the command
			ID: string

			// Human-readable title of the command
			title: string

			// Optional path associated with the command
			path?: string

			// Keyboard shortcuts for the command
			shortcut?: [...string]

			// Keywords associated with the command
			keywords?: [...string]

			// Category for grouping commands
			category?: string
		} @cuetsy(kind="interface")
	}
}]
