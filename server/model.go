package main

// Request Google Structure
type IncomingRequest struct {
	Handler *gHandler `json:"handler,omitempty"`
	Intent  gIntent   `json:"intent,omitempty"`
	Scene   gScene    `json:"scene,omitempty"`
	Session gSession  `json:"session,omitempty"`
	User    gUser     `json:"user,omitempty"`
	Home    gHome     `json:"home,omitempty"`
	Device  gDevice   `json:"device,omitempty"`
	Context gContext  `json:"context,omitempty"`
}
type gHandler struct {
	Name *string `json:"name,omitempty"`
}

type gIntent struct {
	Name   *string       `json:"name,omitempty"`
	Params gIntentParams `json:"params,omitempty"`
	Query  string        `json:"query,omitempty"`
}
type gIntentParams struct {
	Status    *gIntentTimeParameterValue `json:"status,omitempty"`
	Message   *gIntentTimeParameterValue `json:"message,omitempty"`
	OtherUser *gIntentTimeParameterValue `json:"other_user,omitempty"`
}

type gIntentParameterValue struct {
	Original *string `json:"original,omitempty"`
	Resolved string  `json:"resolved,omitempty"`
}

type gIntentTimeParameterValue struct {
	Original *string `json:"original,omitempty"`
	Resolved *string `json:"resolved,omitempty"`
}

type gScene struct {
	Name              *string    `json:"name,omitempty"`
	SlotFillingStatus string     `json:"slotFillingStatus,omitempty"`
	Slots             gSlot      `json:"slots,omitempty"`
	Next              gNextScene `json:"next,omitempty"`
}

// Need to be validated. Leaving with fake "name" property for now.
// Details: https://developers.google.com/assistant/conversational/reference/rest/v1/TopLevel/fulfill#Scene
type gSlot struct {
	Name *string `json:"name,omitempty"`
}

// Need to be validated. Leaving with fake "name" property for now.
// Details: https://developers.google.com/assistant/conversational/reference/rest/v1/TopLevel/fulfill#Scene
type gNextScene struct {
	Name *string `json:"name,omitempty"`
}

type gSession struct {
	ID            *string         `json:"id,omitempty"`
	Params        gSessionParams  `json:"params,omitempty"`
	TypeOverrides []gTypeOverride `json:"typeOverrides,omitempty"`
	LanguageCode  string          `json:"languageCode,omitempty"`
}

// Need to be validated. Leaving with fake "name" property for now.
// Details https://developers.google.com/assistant/conversational/reference/rest/v1/TopLevel/fulfill#Session
type gSessionParams struct {
	Name *string `json:"name,omitempty"`
}

type gTypeOverride struct {
	Name    *string `json:"name,omitempty"`
	Mode    string  `json:"mode,omitempty"`
	Synonym string  `json:"synonym,omitempty"` // Need to be Updated. https://developers.google.com/assistant/conversational/reference/rest/v1/TopLevel/fulfill#SynonymType
}

type gUser struct {
	Locale               *string                    `json:"locale,omitempty"`
	Params               gUserParams                `json:"params,omitempty"`
	AccountLinkingStatus string                     `json:"accountLinkingStatus,omitempty"`
	VerificationStatus   string                     `json:"verificationStatus,omitempty"`
	LastSeenTime         string                     `json:"lastSeenTime,omitempty"`
	Engagement           gUserEngagement            `json:"engagement,omitempty"`
	PackageEntitlements  []gUserPackageEntitlements `json:"packageEntitlements,omitempty"`
}

// Need to be validated. Leaving with fake "name" property for now.
// Details https://developers.google.com/assistant/conversational/reference/rest/v1/TopLevel/fulfill#User
type gUserParams struct {
	Name *string `json:"name,omitempty"`
}
type gUserEngagement struct {
	PushNotificationIntents *string `json:"pushNotificationIntents,omitempty"` // Need to be Updated. https://developers.google.com/assistant/conversational/reference/rest/v1/TopLevel/fulfill#Engagement
	DailyUpdateIntents      string  `json:"dailyUpdateIntents,omitempty"`      // Need to be Updated. https://developers.google.com/assistant/conversational/reference/rest/v1/TopLevel/fulfill#Engagement
}

type gUserPackageEntitlements struct {
	Name *string `json:"name,omitempty"`
}

type gHome struct {
	Params *gHomeParams `json:"params,omitempty"`
}
type gHomeParams struct {
	Name *string `json:"name,omitempty"` // Need to be updated. https://developers.google.com/assistant/conversational/reference/rest/v1/TopLevel/fulfill#Home
}
type gDevice struct {
	Capabilities *[]string `json:"capabilities,omitempty"`
}
type gContext struct {
	Media *gMediaContext `json:"media,omitempty"`
}
type gMediaContext struct {
	Progress *string `json:"progress,omitempty"`
}

// Response Google Structure

type OutgoingResponse struct {
	Prompt   *gPrompt   `json:"prompt,omitempty"`
	Scene    *gScene    `json:"scene,omitempty"`
	Session  gSession   `json:"session,omitempty"`
	User     *gUser     `json:"user,omitempty"`
	Home     *gHome     `json:"home,omitempty"`
	Device   *gDevice   `json:"device,omitempty"`
	Expected *gExpected `json:"expected,omitempty"`
}

type gPrompt struct {
	Override    bool            `json:"override,omitempty"`
	FirstSimple *gSimple        `json:"firstSimple,omitempty"`
	Content     *gContent       `json:"content,omitempty"`
	LastSimple  *gSimple        `json:"lastSimple,omitempty"`
	Suggestions *[]gSuggestions `json:"suggestions,omitempty"`
	Link        *gLink          `json:"link,omitempty"`
	Canvas      *gCanvas        `json:"canvas,omitempty"`
	OrderUpdate *gOrderUpdate   `json:"orderUpdate,omitempty"`
}

type gSimple struct {
	Speech *string `json:"speech,omitempty"`
	Text   string  `json:"text,omitempty"`
}

type gContent struct {
	Card       *gCard       `json:"card,omitempty"`
	Image      *gImage      `json:"image,omitempty"`
	Table      *gTable      `json:"table,omitempty"`
	Media      *gMedia      `json:"media,omitempty"`
	Collection *gCollection `json:"collection,omitempty"`
	List       *gList       `json:"list,omitempty"`
}

type gCard struct {
	Title     *string `json:"title,omitempty"`
	Subtitle  string  `json:"subtitle,omitempty"`
	Text      string  `json:"text,omitempty"`
	Image     gImage  `json:"image,omitempty"`
	ImageFill string  `json:"imageFill,omitempty"`
	Button    string  `json:"button,omitempty"` // Need Updating: https://developers.google.com/assistant/conversational/reference/rest/v1/TopLevel/fulfill#Card
}
type gImage struct {
	URL    *string `json:"url,omitempty"`
	Alt    string  `json:"alt,omitempty"`
	Height int     `json:"height,omitempty"`
	Width  int     `json:"width,omitempty"`
}

type gTable struct {
	Title     *string `json:"title,omitempty"`
	Subtitle  string  `json:"subtitle,omitempty"`
	Image     gImage  `json:"image,omitempty"`
	ImageFill string  `json:"imageFill,omitempty"`
	Columns   string  `json:"columns,omitempty"` // Need Updating: https://developers.google.com/assistant/conversational/reference/rest/v1/TopLevel/fulfill#Table
	Rows      string  `json:"rows,omitempty"`    // Need Updating: https://developers.google.com/assistant/conversational/reference/rest/v1/TopLevel/fulfill#Table
	Button    string  `json:"button,omitempty"`  // Need Updating: https://developers.google.com/assistant/conversational/reference/rest/v1/TopLevel/fulfill#Card
}

type gMedia struct {
	MediaType             *string  `json:"mediaType,omitempty"`
	StartOffset           string   `json:"startOffset,omitempty"`
	OptionalMediaControls []string `json:"optionalMediaControls,omitempty"`
	MediaObjects          string   `json:"mediaObjects,omitempty"` // Need Updating: https://developers.google.com/assistant/conversational/reference/rest/v1/TopLevel/fulfill#Media
}
type gCollection struct {
	Title     *string `json:"title,omitempty"`
	Subtitle  string  `json:"subtitle,omitempty"`
	Items     string  `json:"items,omitempty"` // Need Updating: https://developers.google.com/assistant/conversational/reference/rest/v1/TopLevel/fulfill#Collection
	ImageFill string  `json:"imageFill,omitempty"`
}

type gList struct {
	Title    *string `json:"title,omitempty"`
	Subtitle string  `json:"subtitle,omitempty"`
	Items    string  `json:"items,omitempty"` // Need Updating: https://developers.google.com/assistant/conversational/reference/rest/v1/TopLevel/fulfill#List
}

type gLink struct {
	Name *string  `json:"name,omitempty"`
	Open gOpenURL `json:"open,omitempty"`
}

type gOpenURL struct {
	URL  *string `json:"url,omitempty"`
	Hint string  `json:"hint,omitempty"`
}
type gCanvas struct {
	URL         *string  `json:"url,omitempty"`
	Data        []string `json:"data,omitempty"`
	SuppressMic bool     `json:"suppressMic,omitempty"`
}
type gOrderUpdate struct { //https://developers.google.com/assistant/conversational/reference/rest/v1/TopLevel/fulfill#OrderUpdate
	Type             *string `json:"type,omitempty"`
	Order            string  `json:"order,omitempty"` // Need Updating: https://developers.google.com/assistant/conversational/reference/rest/v1/TopLevel/fulfill#OrderUpdate
	UpdateMask       string  `json:"updateMask,omitempty"`
	UserNotification string  `json:"userNotification,omitempty"` // Need Updating: https://developers.google.com/assistant/conversational/reference/rest/v1/TopLevel/fulfill#OrderUpdate
	Reason           string  `json:"reason,omitempty"`
}

type gExpected struct { //https://developers.google.com/assistant/conversational/reference/rest/v1/TopLevel/fulfill#Expected
	Speech       []string `json:"speech,omitempty"`
	LanguageCode *string  `json:"languageCode,omitempty"`
}

type gSuggestions struct {
	Title string `json:"title"`
}
