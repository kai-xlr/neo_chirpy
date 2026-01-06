package main

const (
	// Chirp constraints
	MaxChirpLength = 140

	// HTTP content types
	ContentTypeJSON      = "application/json"
	ContentTypeTextPlain = "text/plain; charset=utf-8"
	ContentTypeTextHTML  = "text/html; charset=utf-8"

	// Error messages
	ErrMsgDecodeParams     = "Couldn't decode parameters"
	ErrMsgCreateChirp      = "Couldn't create chirp"
	ErrMsgRetrieveChirps   = "Couldn't retrieve chirps"
	ErrMsgMethodNotAllowed = "Method not allowed"
)
