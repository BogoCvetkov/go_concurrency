# Go Concurrent Code Playground

This GitHub repository serves as a playground for exploring concurrent code in Go. The main focus of this project is to practice writing concurrent Go code using goroutines, channels and waitgroups. The repository consists of two projects that showcase different aspects of concurrency.

## Project 1: Pizza Restaurant Simulation

- **Location** -> `/producer-consumer` folder

The first project is a smaller application that simulates a pizza restaurant. It demonstrates the usage of concurrent code to handle pizza orders and processing. The `main.go` file in this project contains all the logic for simulating the pizza restaurant. It utilizes goroutines and channels to manage the pizza order processing. The application tracks the number of successful and failed pizza orders and provides insights into the overall performance of the restaurant.

### Logic of the Pizza Restaurant Simulation

The pizza restaurant simulation uses goroutines and channels to manage the pizza order processing. Here is a brief overview of the logic:

1. The `pizzeria` goroutine runs in the background and attempts to make pizza orders. It calls the `makePizza` function to create each pizza order.

2. The `makePizza` function generates a random number to determine if the pizza can be made successfully. Each pizza has a different production time. If a pizza cannot be made due to certain conditions (e.g., running out of ingredients or the cook quitting), it is marked as failed. Otherwise, it is marked as successfully made.

3. The `Producer` struct holds two channels: one for pizza orders (`data` channel) and another to handle the end of processing (`quit` channel).

4. The main function creates a `Producer` instance and runs the `pizzeria` goroutine in the background.

5. The main function consumes the pizza orders from the `data` channel. It distinguishes between successfully made pizzas and failed ones, displaying appropriate messages. The process continues until the specified number of pizzas is made or the pizza orders are exhausted.


## Project 2: Subscription Web Server -> 

- **Location** -> `/app` folder

The second project is a small web server that simulates a simple application where you can register and buy subscriptions. Concurrency is utilized in this project for sending emails related to user registration and subscription purchase. The web server implements a separate goroutine to handle email sending in the background. Concurrent code is also used when generating PDF invoices and manuals for subscription purchases.

### Logic of the Subscription Web Server

The subscription web server implements a basic user interface for user registration, login, and subscription purchase. Concurrency is used in the `/member/subscribe`, `/register`, and `/login` routes.

Here is a brief overview of the logic:

- The main function creates an instance of the `AppConfig` struct and runs the `listenForEmails` goroutine in the background using `go app.listenForEmails()`.

- The `listenForEmails` function listens for incoming emails on the `app.Mailer.MailChan` channel. It spawns a goroutine to send each email and also listens on the `app.Mailer.DoneChan` & `app.Mailer.ErrChan` channels to log info if the email succeeded/failed

- The `shutGracefully` function is responsible for graceful server shutdown. It listens for interrupt signals and uses a waitgroup to wait for all running goroutines (incremented by concurrent processes like email sending and PDF generation) to finish before shutting down the server.

- The web server provides server-side rendering of the user interface, allowing users to register, login, and choose from available subscriptions for purchase. Concurrency is implemented in the `/member/subscribe` route for generating PDF invoices and manuals and sending them via the email channel. Concurrent code is also used in the `/register` and `/login` routes to handle failed login attempts.

