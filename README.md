# mercurio
Bare Notification Service based in Golang and SSE.

## Okay, so what is it actually?

This is a prototype of a [Notification Service](https://en.wikipedia.org/wiki/Notification_service) (in the vein of what you get while using YouTube/Facebook/LinkedIn and the likes) that leverages [server-sent events](https://developer.mozilla.org/en-US/docs/Web/API/Server-sent_events) to deliver one way communication in a quick and safe manner. Any time there is an event on the server site, it is pushed to the client near real time. It supports *unicast* (one-to-one) and *broadcast* (one-to-many) models of event notification. And there is also an API where client can fetch previous notifications and stuff.

For security, it uses [JWT](https://jwt.io/) -- even on the SSE channel (a.k.a. [EventSource](https://developer.mozilla.org/en-US/docs/Web/API/EventSource)). In order to pass custom HTTP headers, I've got [Viktor's EventSource Polyfill](https://github.com/Yaffle/EventSource/) in the train.

As it is a prototype, [SQLite](https://www.sqlite.org/index.html) is being used for persistence. To make it even easier, [GORM](https://gorm.io/) is in charge of migrations and object-relational mapping.

There is also an optional feature (turned `off` by default) of using an underlying [RabbitMQ](https://www.rabbitmq.com/) to allow for horizontal scaling. It could well being [ActiveMQ](https://activemq.apache.org/), [Amazon SQS](https://aws.amazon.com/sqs/), [Redis Pub/Sub](https://redis.io/topics/pubsub) or whatever message-oriented middleware platform for that matter. I went with RabbitMQ because I got it running on Docker container so why not?

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

If you'd like to get your hands dirt with this little project, there is always somenthing that could be done. Send me a pull request if like. And of course, feel free to reach me out to discuss any feature or anything else.

Tasks that I have in mind now are as follow:

* Write a Dockerfile;
* Definitely improve logging;
* Implement a nice client side app to show case a full closed loop;
* Grow up and better the automated test suite (it's quite poor yet);
* Clients might be interested in receive notifications of a certain sort;
* Automate GitHub pipeline;
* Add HTTPS (maybe -- 'cause we can just have that on load balancer).

## Copyright

Leandro Silva <leandrosilva.com.br>