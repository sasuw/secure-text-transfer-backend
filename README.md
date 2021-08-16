<!-- ABOUT THE PROJECT -->
## About Secure Text Transfer

Secure Text Transfer (STT) is a website project enabling people to transfer a piece of text (e.g. a password, URL, text snippet) conveniently and securely from one device to another. Design goals:

  * no registration or login needed
  * minimal amount of cognitive load
  * minimal amount of typing and clicks without sacrifing security
  * responsive design, wide device support

[![STT Screenshot](https://github.com/sasuw/secure-text-transfer-backend/blob/master/images/stt-fronpage-screenshot-2020-08-30.png?raw=true)](https://stt.sasu.net)

### History

I started this project in 2020, because over the years I often had the need to quickly copy a password from one device to another. As secure passwords are quite long nowadays and the amount of typing can be quite frustrating, I thought there should be a way to copy a text string from one trusted device to another securely. As I did not find any ready-made solution fitting my needs, I decided to make it myself.

### Security

Currently the text is transferred from the sending device to the server, stored in working memory for a maximum time of 5 minutes or until it is retrieved and then deleted. It is not written anywhere permanently, whether in a database nor is it logged anywhere. Transport security is guaranteed by using HTTPS.

In the future the text to be transferred is encrypted on the client side, further increasing the security.

### Project structure

The project has a frontend consisting of a one-page static website, see [secure-text-frontend](https://github.com/sasuw/secure-text-transfer-frontend). The frontend interacts with the backend through REST endpoints using AJAX.

This project is the backend part, which is a web server with REST endpoints created with Golang.

### Built With

* [Mux](https://github.com/gorilla/mux)

<!-- GETTING STARTED -->
## Getting Started

### Prerequisites

To run the backend locally, you should have a working go environment and a text editor or an IDE with a built-in go environment.

### Build and run

<u>Prepare</u>

    go mod init sttme.net/backend
    go mod tidy

<u>Build</u>

    go build main.go limit.go jsonHandler.go

<u>Run</u>

    ./main

Running the main program without arguments starts the server on port 9999. If you want to run it using a different port, supply the environment variable STT_PORT, e.g. like this

    export STT_PORT=9998 && ./main

Because browsers don't allow cross-site AJAX requests willy-nilly, you may have to provide the environment variable STT_ENV=dev when testing locally to prevent errors in the frontend, like this

    export STT_ENV=dev && ./main

<u>Test</u>

When testing the API, you must provide a

    X-Requested-With

header with value

    XMLHttpRequest

with every request.

## API Documentation

### Methods

<u>POST /string</u>

Submits string to be stored on the server until it is retrieved. If string is not retrieved within five minutes, it is deleted from the server.

Return codes and values:

  * 200 OK: when request is processed successfully. Returns the PIN with the value of 1 to 99999 as a string.
  * 204 No content: when posted string is empty
  * 413 Payload too large: when string length is over 4000 characters
  * 503 Service unavailable: when PIN could not be generated for some reason

<u>POST /pin</u>

Submits PIN to retrieve previously stored string from server.

Return codes and values:

  * 200 OK: when request is processed successfully and stored string is found with PIN. Returns the stored string and deletes it from the server.
  * 204 No content: no string is found with PIN
  * 503 Service unavailable: internal error

### General errors

429 Too Many Requests: when too many requests are coming from the same IP address too often


<!-- ROADMAP -->
## Roadmap

See the [open issues](https://github.com/sasuw/secure-text-transfer-backend/issues) for a list of proposed features (and known issues).

<!-- CONTRIBUTING -->
## Contributing

You can contribute to this project in many ways:

  * submitting an issue (bug or enhancement proposal) 
  * testing
  * contributing code

If you want to contribute code, please open an issue or contact me beforehand to ensure that your work in line with the project goals.

When you decide to commit some code:

1. Fork the Project
2. Create your Feature Branch (`git checkout -b feature/AmazingFeature`)
3. Commit your Changes (`git commit -m 'Add some AmazingFeature'`)
4. Push to the Branch (`git push origin feature/AmazingFeature`)
5. Open a Pull Request


<!-- LICENSE -->
## License

Distributed under the MIT License. See `LICENSE` for more information.


<!-- CONTACT -->
## Contact

Sasu Welling - [@sasuw](https://twitter.com/sasuw) -  
Project Link: [https://github.com/sasuw/secure-text-transfer-backend](https://github.com/sasuw/secure-text-transfer-backend)



<!-- ACKNOWLEDGEMENTS -->
## Acknowledgements
* [Mux](https://github.com/gorilla/mux)
* [decodeJSONBody](https://www.alexedwards.net/blog/how-to-properly-parse-a-json-request-body)
* [limit.go](https://www.alexedwards.net/blog/how-to-rate-limit-http-requests)
