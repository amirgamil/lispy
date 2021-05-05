# Carly ✉️
#### Genereate beautiful letters for your loved ones that can be shared in seconds
Carly was written to help me send a birthday letter to my mum halfway around the world. Now anyone can do it too!


![letter1](https://user-images.githubusercontent.com/7995105/116254324-3fcb5300-a73f-11eb-8398-fcf53bd9d1dc.png)
![letter2](https://user-images.githubusercontent.com/7995105/116254322-3f32bc80-a73f-11eb-8f33-059b34550093.png)
![letter3](https://user-images.githubusercontent.com/7995105/116254321-3f32bc80-a73f-11eb-955c-64624f298e41.png)
![letter4](https://user-images.githubusercontent.com/7995105/116254320-3f32bc80-a73f-11eb-984d-0e01913439f5.png)

## Details
Carly is written in Next.js + React on the frontend (hosted on Vercel) and Go with MongoDB Atlas on the backend (hosted as a systemd file with nginx on Digital Ocean).

To run this locally, you will need to run the web server in Go (using `go run main.go`) as well as the frontend (navigate inside the frontend directory and run `yarn dev`).

You will also need a MongoDB atlas account (free tier M0 can be accessed by anyone). Once you've made an account, configure a user with admin access and store the username, password, and shared URL in a .env file with MONGO_USER, MONGO_PASS, MONGO_SHARD_URL as the variable names respectively. Make sure that this .env file is located in the same directory as the go.mod file.

You will also need to create a .env.local file inside the frontend folder and populate it with two variables 
`NEXT_PUBLIC_HOST=localhost:3000
NEXT_PUBLIC_HOSTAPI=127.0.0.1:port/api`
You can select any port like 8998. 

## API
The API provides two endpoints 
### `POST /api`
- Accepts a JSON body that looks like this:
`{
  "title": "titleLetter",
  "expiry": "getExpiryDate()", 
  "password": "",
  "content": [
    {
    "person": "person1",
    "msg": "msg1",
    "imgAdd": "imgAdd1"
   },
   {
     "person":"person2",
     "msg": "msg2",
     "imgAdd": "imgAdd2"
   }...
  ]
}`

The API returns the genereated hash if successful
`#200 OK
#{ "hash": "166989a"}`

### `GET /api/{hash}`
The API returns
`#401 Unauthorized` if the letter is password protected

OR 

`#404 Bad request` if it does not find the hash in the database (or for any other possible error)

## Contributing

Contributions are what make the open source community such an amazing place to be learn, inspire, and create. Any contributions you make are **greatly appreciated**.

1. Fork the Project
2. Create your Feature Branch (`git checkout -b feature/AmazingFeature`)
3. Commit your Changes (`git commit -m 'Add some AmazingFeature'`)
4. Push to the Branch (`git push origin feature/AmazingFeature`)
5. Open a Pull Request


## Acknowledgements
* [Ctrl-v](https://github.com/jackyzha0/ctrl-v)
* [Block CSS](https://thesephist.github.io/blocks.css/)
