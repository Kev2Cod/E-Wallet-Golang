package controller

import (
	Model "dumbways/model"
	"encoding/json"
	"fmt"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/gofiber/fiber"
	"github.com/streadway/amqp"
)

const jwtSecret = "secret"
const url = "amqp://guest:guest@localhost:5672/"

var wallet []Model.Wallet
var transactions []Model.Transaction
// var errors []Model.ErrorMessage


// ========== Top up Consume ===========
func ConsumeTopUp(ctx *fiber.Ctx) {
	conn, err := amqp.Dial(url)

	if err != nil {
		fmt.Println("❌ Failed Initializing Broker Connection")
		panic(err)
	}

	ch, err := conn.Channel()
	if err != nil {
		fmt.Println(err)
	}
	defer ch.Close()

	if err != nil {
		fmt.Println(err)
	}

	msgs, err := ch.Consume(
		"top_up",
		"",
		true,
		false,
		false,
		false,
		nil,
	)

	if err != nil {
		fmt.Println(err)
	}

	fmt.Println("Successfully Connected to our RabbitMQ Instance 🎉")
	fmt.Println(" [*] - Waiting for messages Top up")

	forever := make(chan bool)

	go func() {
		for d := range msgs {

			if len(wallet) == 0 {
				wallet = append(wallet, Model.Wallet{Balance: 0})
			}

			var jsonData = []byte(d.Body)
			var data Model.TopUp

			var err = json.Unmarshal(jsonData, &data)
			if err != nil {
				fmt.Println(err.Error())
				return
			}

			s := wallet[len(wallet)-1]
			
			saldo := s.Balance + data.Amount

			fmt.Printf("\n ✅ Top up 💸 Success 👍 \n")
			// Save balance 
			wallet = append(wallet, Model.Wallet{Balance: saldo})
			fmt.Printf("\n 💰 Balance : %f\n", saldo)

			// fmt.Println(wallet)
		}
	}()
	<-forever
}


// ========== Transaction Consume ===========

func ConsumeWallet(ctx *fiber.Ctx) {
	conn, err := amqp.Dial(url)

	if err != nil {
		fmt.Println("❌ Failed Initializing Broker Connection")
		panic(err)
	}

	ch, err := conn.Channel()
	if err != nil {
		fmt.Println(err)
	}
	defer ch.Close()

	if err != nil {
		fmt.Println(err)
	}

	msgs, err := ch.Consume(
		"e_wallet",
		"",
		true,
		false,
		false,
		false,
		nil,
	)

	if err != nil {
		fmt.Println(err)
	}

	fmt.Println("Successfully Connected to our RabbitMQ Instance 🎉")
	fmt.Println(" [*] - Waiting for messages")

	forever := make(chan bool)

	go func() {

		for d := range msgs {

			var jsonData = []byte(d.Body)
			var data Model.Transaction

			var err = json.Unmarshal(jsonData, &data)
			if err != nil {
				fmt.Println(err.Error())
				return
			}

				if len(wallet) == 0 {
					wallet = append(wallet, Model.Wallet{Balance: 0})
				}

				s := wallet[len(wallet)-1]

				// Kondisi untuk pengecekan
				if s.Balance == 0 || s.Balance < data.Price {
					fmt.Printf(" ❌ Maaf Sisa Saldo anda Tidak Mencukupi ❌ \n")
				} else{
				saldo := s.Balance - data.Price

				fmt.Println(saldo)

				wallet = append(wallet, Model.Wallet{Balance: saldo})

				// print sisa saldo
				fmt.Printf("\n ✅ Payment Success ❕ \n")
				fmt.Printf("\n ➡️  Balance : %f \n", saldo)
				// Simpan data transaksi
				transactions = append(transactions, Model.Transaction(data))
				}

		}


	}()
	<-forever
}

// ========== Topup Publish ===========
func Topup(ctx *fiber.Ctx) {
	
	var body Model.TopUp
	err := ctx.BodyParser(&body)
	if err != nil {
		fmt.Println("Please input data")
		panic(err)
	}

	x := Model.TopUp{
		Amount : body.Amount,
	}

	// data akan dikirim ke message broker
	data, _ := json.Marshal(x)

	// RabbitMQ
	conn, err := amqp.Dial(url)
	if err != nil {
		fmt.Println("❌ Failed Initializing Broker Connection")
		panic(err)
	}

	ch, err := conn.Channel()
	if err != nil {
		fmt.Println(err)
	}
	defer ch.Close()

	q, err := ch.QueueDeclare(
		"top_up",
		false,
		false,
		false,
		false,
		nil,
	)

	fmt.Println(q)

	if err != nil {
		fmt.Println(err)
	}

	err = ch.Publish(
		"",
		"top_up",
		false,
		false,
		amqp.Publishing{
			ContentType:  "application/json",
			Body:         data,
			DeliveryMode: amqp.Persistent,
		},
	)
	if err != nil {
		fmt.Println(err)
	}

	ctx.Send("✅ Berhasil Menambahkan data Top up ke Queue")
}

func Subscribe(ctx *fiber.Ctx) {
	var body Model.Transaction
	err := ctx.BodyParser(&body)
	if err != nil {
		fmt.Println("❌ Please Enter data correctly")
		panic(err)
	}

	x := Model.Transaction{
		Messages: body.Messages,
		Price: body.Price,
	}

	data, _ := json.Marshal(x)

	conn, err := amqp.Dial(url)
	if err != nil {
		fmt.Println("❌ Failed Initializing Broker Connection")
		panic(err)
	}

	ch, err := conn.Channel()
	if err != nil {
		fmt.Println(err)
	}
	defer ch.Close()

	q, err := ch.QueueDeclare(
		"e_wallet",
		false,
		false,
		false,
		false,
		nil,
	)

	fmt.Println(q)

	if err != nil {
		fmt.Println(err)
	}

	err = ch.Publish(
		"",
		"e_wallet",
		false,
		false,
		amqp.Publishing{
			ContentType:  "application/json",
			Body:         data,
			DeliveryMode: amqp.Persistent,
		},
	)
	if err != nil {
		fmt.Println(err)
	}

	ctx.Send("✅ Berhasil Menambahkan data Transaksi ke Queue 🎉")
}



// ============= Login ===============
func Login(ctx *fiber.Ctx) {

	type request struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	var body request
	err := ctx.BodyParser(&body)

	if err != nil {
		ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "cannot parse json",
		})
		return
	}

	if body.Email != "user@mail.com" || body.Password != "123" {
		ctx.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "❌ Bad Credentials",
		})
		return
	}

	token := jwt.New(jwt.SigningMethodHS256)
	claims := token.Claims.(jwt.MapClaims)
	claims["id"] = 1
	claims["name"] = "Kevin"
	claims["exp"] = time.Now().Add(time.Hour * 24 * 7) // a week

	s, err := token.SignedString([]byte(jwtSecret))
	if err != nil {
		ctx.SendStatus(fiber.StatusInternalServerError)
		return
	}

	ctx.Status(fiber.StatusOK).JSON(fiber.Map{
		"token": s,
		"user": struct {
			Name  string `json: "name"`
			Email string `json:"email"`
		}{
			Name: "Kevin",
			Email: body.Email,
		},
	})
}

// ============ Wallet =============
func Wallet(ctx *fiber.Ctx) {
	user := ctx.Locals("user").(*jwt.Token)
	claims := user.Claims.(jwt.MapClaims)
	name := claims["name"].(string)

	if len(wallet) == 0 {
		wallet = append(wallet, Model.Wallet{Balance: 0})
	}

	 w := wallet[len(wallet)-1]

	ctx.Status(fiber.StatusOK).JSON(fiber.Map{
		"user": struct {
			Name string `json:"name"`
		}{
			Name: name,
		},
		"balance": w.Balance,
		"transactions": transactions,
	})
}
