start1:
	go run . --listen :1000 --ctl :1001

start2:
	go run . --listen :2000 --ctl :2001 --contact localhost:1000

start3:
	go run . --listen :3000 --ctl :3001 --contact localhost:2000

start4:
	go run . --listen :4000 --ctl :4001 --contact localhost:3000

start5:
	go run . --listen :5000 --ctl :5001 --contact localhost:4000
