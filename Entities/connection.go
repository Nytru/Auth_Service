package entities

type RefreshToken struct {
	Token 	  string `json:"token"`
	ExpiresAt int64	 `json:"exp"`
}

type User struct {
	Value 		  string 	   `json:"val,omitempty"`
	GUID 		  string 	   `json:"guid"`
	Refreshtoken  RefreshToken `json:"reftoken"`
}
