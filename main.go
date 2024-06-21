package main

import (
	"fmt"
	"github.com/mozillazg/go-pinyin"
	"math/rand"
	"os"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"test-src/rand_stuff"
	"time"
)

var wg sync.WaitGroup
var sem = make(chan struct{}, runtime.NumCPU()*4) // 控制并发数

var sqlUser = `
DROP TABLE IF EXISTS users;
-- 用户表：存储用户的基本信息
CREATE TABLE users (
    id INT AUTO_INCREMENT PRIMARY KEY COMMENT '用户的唯一标识符',
    usercode VARCHAR(50) NOT NULL UNIQUE COMMENT '用户名，必须唯一',
    password VARCHAR(255) NOT NULL COMMENT '用户密码，经过加密存储',
    email VARCHAR(100) NOT NULL UNIQUE COMMENT '用户的电子邮件地址，必须唯一',
    username VARCHAR(100) COMMENT '用户的全名',
    phone VARCHAR(11) COMMENT '用户的手机号',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '记录用户创建的时间',
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '记录用户最后更新的时间'
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='用户表：存储用户的基本信息';
`

var sqlProducts = `
DROP TABLE IF EXISTS products;
-- 产品表：存储产品的基本信息
CREATE TABLE products (
    id INT AUTO_INCREMENT PRIMARY KEY COMMENT '产品的唯一标识符',
    name VARCHAR(100) NOT NULL COMMENT '产品名称',
    descr TEXT COMMENT '产品描述',
    price DECIMAL(10, 2) NOT NULL COMMENT '产品价格',
    stock INT(100) DEFAULT 0 COMMENT '产品库存数量',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '记录产品创建的时间',
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '记录产品最后更新的时间'
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='产品表：存储产品的基本信息';
`

var sqlOrder = `
DROP TABLE IF EXISTS orders;
-- 订单表：存储用户的订单信息
CREATE TABLE orders (
    id INT AUTO_INCREMENT PRIMARY KEY COMMENT '订单的唯一标识符',
    user_id INT NOT NULL COMMENT '关联到用户表的用户ID',
    product_id INT NOT NULL COMMENT '关联到产品表的产品ID',
    quantity INT NOT NULL COMMENT '产品的数量',
    price DECIMAL(10, 2) NOT NULL COMMENT '产品的单价',
    total_amount DECIMAL(10, 2) NOT NULL COMMENT '订单的总金额',
    status INT NOT NULL COMMENT '订单的状态(0:待支付,1:完成支付,2:待发货,3:已发货,4:已送达,5:已签收)',
    order_date TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '订单创建的时间'
    -- FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE ON UPDATE CASCADE,
	-- FOREIGN KEY (product_id) REFERENCES products(id) ON DELETE CASCADE ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='订单表：存储用户的订单信息';
`

// 全局递增ID
var userID = 0
var productID = 1

// GenSQL struct
type GenSQL struct{}

func (gs *GenSQL) strTimeProp(start, end string, prop float64, layout string) int64 {
	startTime, _ := time.Parse(layout, start)
	endTime, _ := time.Parse(layout, end)
	ptime := startTime.Unix() + int64(prop*float64(endTime.Unix()-startTime.Unix()))
	return ptime
}

func (gs *GenSQL) randomTimestamp(layout string) int64 {
	start := "2016-06-02 12:12:12"
	end := "2024-07-27 00:00:00"
	return gs.strTimeProp(start, end, rand.Float64(), layout)
}

func (gs *GenSQL) createPhone() string {
	prelist := []string{"130", "131", "132", "133", "134", "135", "136", "137", "138", "139",
		"147", "150", "151", "152", "153", "155", "156", "157", "158", "159",
		"186", "187", "188", "189"}
	pre := prelist[rand.Intn(len(prelist))]
	phone := pre
	for i := 0; i < 8; i++ {
		phone += fmt.Sprintf("%d", rand.Intn(10))
	}
	return phone
}

func inserProducts(productsFile *os.File) {
	var productRowStr []string
	for i, s := range rand_stuff.GetGoods() {
		name := fmt.Sprintf("%v", s)
		description := fmt.Sprintf("产品描述: 非常好用的%v", s)
		price := float64(rand.Intn(10000)) / 100.1 // 随机价格
		stock := rand.Intn(100000000) + 9000000    // 随机库存
		timestamp := time.Now().Format("2006-01-02 15:04:05")
		productRow := fmt.Sprintf("(%d,'%s','%s',%.2f,%d,'%s','%s')", i+1, name, description, price, stock, timestamp, timestamp)
		productRowStr = append(productRowStr, productRow)
		productID++
	}
	productInsertStr := fmt.Sprintf("INSERT IGNORE INTO products(id, name, descr, price, stock, created_at, updated_at) VALUES %s;\n", strings.Join(productRowStr, ","))
	productsFile.WriteString(productInsertStr)
}

func inserOrder(gs *GenSQL, RandName string, userFile, orderFile *os.File, uID int) {
	var userRowStr string
	var orderRowStr []string

	curTime := time.Now()

	timestamp := gs.randomTimestamp("2006-01-02 15:04:05")
	createTime := time.Unix(timestamp, 0).Format("2006-01-02 15:04:05")

	hanz := pinyin.NewArgs()
	Piny := pinyin.Pinyin(RandName, hanz)
	usercode := ""
	for _, s := range Piny {
		usercode = usercode + s[0]
	}
	//usercode = usercode + "_" + strconv.Itoa(rand.Intn(99999))
	phone := gs.createPhone()

	userRow := fmt.Sprintf("(%d,'%s','%s','%s','%s','%s','%s','%s')", uID, usercode, "******", fmt.Sprintf("%v@example.com", usercode), RandName, phone, createTime, createTime)

	nums := rand.Intn(200)

	for i := 0; i < nums; i++ {
		randomOrderCount := rand.Intn(4) + 1
		if randomOrderCount > 0 {
			for c := 0; c < randomOrderCount; c++ {
				timestamp = gs.randomTimestamp("2006-01-02 15:04:05")
				orderCreateTime := time.Unix(timestamp, 0).Format("2006-01-02 15:04:05")
				//orderUpdateTime := time.Unix(timestamp, 0)+1.Format("2006-01-02 15:04:05")
				orderRow := fmt.Sprintf("(%d,%d,%d,%d,%.2f,%d,'%s')", uID, rand.Intn(productID)+1, rand.Intn(5)+1, rand.Intn(199)+10, float64(rand.Intn(199)+10)*1.2, rand.Intn(6), orderCreateTime)
				orderRowStr = append(orderRowStr, orderRow)
			}
		}
	}

	userRowStr = fmt.Sprintf("INSERT IGNORE INTO users(id, usercode, password, email, username, phone, created_at, updated_at) VALUES %s;\n", userRow)
	order := fmt.Sprintf("INSERT IGNORE INTO orders(user_id, product_id, quantity, price, total_amount, status, order_date) VALUES %s;\n", strings.Join(orderRowStr, ","))
	fmt.Println("User:", RandName, "Order:", nums, "Time:", time.Now().Sub(curTime))
	userFile.WriteString(userRowStr)
	if nums != 0 {
		orderFile.WriteString(order)
	}

	wg.Done()
	<-sem
}

func (gs *GenSQL) GenerateSqlData() {
	err := os.MkdirAll("sql", 0755)
	if err != nil {
		fmt.Println(err)
		return
	}

	userFile, err := os.Create("sql/test_users_go.sql")
	defer userFile.Close()
	if err != nil {
		fmt.Println(err)
		return
	}

	orderFile, _ := os.Create("sql/test_orders_go.sql")
	defer orderFile.Close()

	productsFile, _ := os.Create("sql/test_products_go.sql")
	defer productsFile.Close()

	_, err = userFile.WriteString(sqlUser)
	if err != nil {
		fmt.Println(err)
		return
	}
	orderFile.WriteString(sqlOrder)
	productsFile.WriteString(sqlProducts)

	inserProducts(productsFile)

	if len(os.Args) < 2 {
		fmt.Println("请输入纯数字,将取高于80%的随机数作为用户数!")
		return
	}

	userNums, err := strconv.Atoi(os.Args[1])
	if err != nil {
		fmt.Println("请输入纯数字,将取高于80%的随机数作为用户数!")
		return
	}
	lowLine := int(float64(userNums) * 0.8)

	for x := 0; x < rand.Intn(userNums)+lowLine; x++ {
		sem <- struct{}{}
		wg.Add(1)
		RandName := rand_stuff.GenRandName()
		userID++
		go inserOrder(gs, RandName, userFile, orderFile, userID)
	}
	wg.Wait()
}

func main() {
	startTime := time.Now()
	fmt.Println("开始时间", startTime)

	gsd := GenSQL{}
	gsd.GenerateSqlData()

	endTime := time.Now()
	fmt.Println("结束时间", endTime, "共持续", endTime.Sub(startTime).Seconds(), "秒")
}
