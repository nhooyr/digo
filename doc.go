/*
Provide a low level interface to the Discord API but also maintain clean readable code.
TODO think about the license
TODO need any nulls?
TODO update test names
TODO need context.Context on rest API methods?
TODO no race conditions in gateway.go and it implements everything correctly?
TODO which events need to be public?
TODO https://discordapp.com/developers/docs/topics/oauth2#get-current-application-information
TODO sharding
TODO do not modify any event data in handlers
TODO why are presence updates given an entire user object? why not just a user id?
TODO presence update for user that was originally offline when bot joined a large guild? Will it contain all fields?
TODO no message state tracking maybe?
TODO use map everywhere except for messages, and messages slice will have very low capacity.
TODO am I overusing sync.RWMutex? Benchmark maybe.

RestAPI does not use One Methods e.g. Channels().One(cID) because too much stutter.
RestAPI does not have all methods defined on single Client because too little structure and hard code completion/doc finding
RestAPI Client methods follow Discord API naming as close as possible ignoring the HTTP method names. HTTP method names do not always get the point across on what is happening. E.g. POST on Messages().BulkDelete or TriggerTypingIndicator
RestAPI Client must be discoverable
RestAPI Client must be easy to use and easy to read the resulting code (more important to have good readability and 0 magic
*/
package discgo
