This is Go (golang) client library for Google Identity Toolkit services.

Thanks to the Chinese GFW, I have problem access to google API. So, the code is not test yet.(This statement will be remove after I have the code test)

##### Sample usage

see cmd/example

##### Misc

When obtaining a key from the Google API console it will be downloaded in a PKCS12 encoding. To use this key you will need to convert it to a PEM file. This can be achieved with openssl. And put path of the output file to gitkit-server-config.json serviceAccountPrivateKeyFile.
```bash
openssl pkcs12 -in key.p12 -nocerts -passin pass:notasecret -nodes -out key.pem
```
