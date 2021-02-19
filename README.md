# mercurio
Bare Notification Service based in Golang and SSE.

## Okay, so what is it actually?

This is a prototype of a [Notification Service](https://en.wikipedia.org/wiki/Notification_service) (in the vein of what you get while using YouTube/Facebook/LinkedIn and the likes) that leverages [server-sent events](https://developer.mozilla.org/en-US/docs/Web/API/Server-sent_events) to deliver one way communication in a quick and safe manner. Any time there is an event on the server site, it is pushed to the client near real time. It supports *unicast* (one-to-one) and *broadcast* (one-to-many) models of event notification. And there is also an API where client can fetch previous notifications and stuff.

For security, it uses [JWT](https://jwt.io/) -- even on the SSE channel (a.k.a. [EventSource](https://developer.mozilla.org/en-US/docs/Web/API/EventSource)). In order to pass custom HTTP headers, I've got [Viktor's EventSource Polyfill](https://github.com/Yaffle/EventSource/) in the train.

As it is a prototype, [SQLite](https://www.sqlite.org/index.html) is being used for persistence. To make it even easier, [GORM](https://gorm.io/) is in charge of migrations and object-relational mapping.

# What is included?

* Source code is in `./src` folder;
* There is some manual tests on `./test` folder;
* While we're in the subject of tests, there is little automated tests yet;
* In the folder `auth` there is files quite utils in dev/test time: `private-key` (oh so super secret) and `test-tokens.txt` (a few valid JWT tokens);
* `.env` files as you'd like.

# Where to go from here?

Well, production would be a nice place to land on. Even though **it's not ready for production yet**, it is quite in its way.

I've let things fairly well arranged and malleable **for a prototype**, with `.env` files, [separation of concerns](https://en.wikipedia.org/wiki/Separation_of_concerns) and stuff, in such a way that "not much" has to be done before a beta candidate to production, I guess.

## TODO

If you'd like to get your hands dirt with this little project, there is always somenthing that could be done. Send me a pull request if like.

* Write a Dockerfile;
* Definitely improve log;
* Grow up and better the automated test suite (it's quite poor yet);
* Automate GitHub pipeline;
* Add HTTPS (maybe -- 'cause we can just have that on load balancer).

## Copyright

Leandro Silva <leandrosilva.com.br>