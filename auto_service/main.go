package main

import (
    "fmt"
    "github.com/gin-gonic/gin"
    "gorm.io/driver/sqlite"
    "gorm.io/gorm"
    "net/http"
)

type ServiceRecord struct {
    ID          uint    `gorm:"primaryKey"`
    Date        string  `json:"date" binding:"required"`
    Description string  `json:"description" binding:"required"`
    Cost        float64 `json:"cost" binding:"required"`
    CarID       uint    `json:"car_id"`
}

type Car struct {
    ID      uint          `gorm:"primaryKey"`
    Model   string        `json:"model" binding:"required"`

    Year    int           `json:"year" binding:"required"`
    VIN     string        `json:"vin" binding:"required"`
    Records []ServiceRecord `json:"records" gorm:"foreignKey:CarID"`
}

var db *gorm.DB

func main() {
    var err error
    db, err = gorm.Open(sqlite.Open("auto_service.db"), &gorm.Config{})
    if err != nil {
        panic("Ошибка открытия базы данных")
    }

    db.AutoMigrate(&Car{}, &ServiceRecord{})

    router := gin.Default()
 
 router.Static("/static", "./static") // Обслуживание статических файлов

    router.LoadHTMLGlob("templates/*")
    router.GET("/", viewCars)
    router.POST("/add", addServiceRecord)

    router.Run(":8080")
}

func viewCars(c *gin.Context) {
    var cars []Car
    if err := db.Preload("Records").Find(&cars).Error; err != nil {
        c.String(http.StatusInternalServerError, "Ошибка при извлечении данных: %s", err.Error())
        return
    }

    c.HTML(http.StatusOK, "index.html", gin.H{
        "cars": cars,
    })
}

func addServiceRecord(c *gin.Context) {
    model := c.PostForm("model")
    yearStr := c.PostForm("year")
    vin := c.PostForm("vin")

    var year int
    _, err := fmt.Sscanf(yearStr, "%d", &year)
    if err != nil {
        fmt.Println("Ошибка преобразования года:", err)
        c.HTML(http.StatusBadRequest, "index.html", gin.H{"error": "Ошибка в году: " + err.Error()})
        return
    }

    var car Car
    car.Model = model
    car.Year = year
    car.VIN = vin

    record := ServiceRecord{
        Date:        c.PostForm("date"),
        Description: c.PostForm("description"),
    }

    costStr := c.PostForm("cost")
    var cost float64
    if _, err := fmt.Sscanf(costStr, "%f", &cost); err != nil {
        c.HTML(http.StatusBadRequest, "index.html", gin.H{"error": "Ошибка в стоимости: " + err.Error()})
        return
    }
    record.Cost = cost

    found := false
    if err := db.Where("vin = ?", car.VIN).First(&car).Error; err == nil {
        record.CarID = car.ID
        car.Records = append(car.Records, record)
        db.Save(&car)
        found = true
        fmt.Println("Обновлена запись для VIN:", car.VIN)
    }

    if !found {
        record.CarID = 0
        car.Records = []ServiceRecord{record}
        db.Create(&car)
        fmt.Println("Создана новая запись для VIN:", car.VIN)
    }

    c.Redirect(http.StatusFound, "/")
}


