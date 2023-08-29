# Pretengo
Basic Account Server Emulator Returning Static Response just enough for Cemu to Play MHF-Z

## Pretengo config.json
Pretengo configuration `config.json`:
```
{
  "ListenAddress": "127.0.0.1",
  "ListenPort": "443",
  "access_token": "1234567890abcdef1234567890abcdef",
  "refresh_token": "fedcba0987654321fedcba0987654321fedcba12",
  "expires_in": "3600",
  "StaticKey": "no",
  "SignKey": "U0VSVklDRVNFUlZJQ0VTRVJWSUNFU0VSVklDRVNFUlZJQ0VTRVJWSUNFU0VSVklDRVNFUlZJQ0VTRVI="
}
```
- When starting MHF on Cemu, Pretengo will listen and send the static response based on config above and using the SessionKey as SignIn Method for Erupe server.
if the StaticKey is set to "no", the SignKey will be generated automatically from userID and password based on Cemu online files.
- The default port used by nintendo is 443 or SSL port, you can change it but you must proxy the 443 into the destination port in order to function.

## How to Use
1. first you need to redirect nintendo.account.net into your local IP wich pretengo listen to, easiest way is to use host file in %windir%/System32/Drivers/etc/hosts and add the following e.g. : ```127.0.0.1 account.nintendo.net```
3. second start the pretengo using go command ```go run pretengo.go```
4. run Cemu and see if pretengo get a post request and can answer with static responde in config.json file

## Note
- the serverlist used by MHF wii-u is srv-mhf-wiiu.capcom-networks.jp you need to redirect this into zerulight.cc or host your own to http server in order to play MHF on cemu
- inspired from https://github.com/SmmServer/pretendoplusplus
- Use latest Erupe server as it support Z2 (wii-u latest version of mhf) https://github.com/ZeruLight/Erupe
