# Part of an authentication service
 
Serving 2 routes
/getnewtokens : should contains query params in url in /getnewtokens?guid={your_guid}&value={your_value}
/refreshtoken : can be easily accessed after getting new tokens. No query requers. Tokens store in cookie
./env/.env - folder with envs. Should contain "KEY" (sha512 encrypting key), "DB_FULL_PASS" (mongo connecting path), "ACCESS_DURATION" and "REFRESH_DURATION" values (duration of tokens in nanosecond 9e11 equals to 15 min)

