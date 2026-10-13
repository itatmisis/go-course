package main

func main() {
	app := buildApp()
	defer app.Close()

	app.Logger.Fatal(app.Start(":8080"))
}
