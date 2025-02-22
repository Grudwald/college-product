package main

import (
    "net/http"
    "strconv"

    "github.com/gin-gonic/gin"
    "gorm.io/driver/sqlite"
    "gorm.io/gorm"
)

type ServiceRecord struct {
    ID          uint    `gorm:"primaryKey"` // Уникальный идентификатор записи обслуживания
    Date        string  `json:"date" binding:"required"`
    Description string  `json:"description" binding:"required"`
    Cost        float64 `json:"cost" binding:"required"`
    CarID       uint    `json:"car_id"`
}

type Car struct {
    ID      uint           `gorm:"primaryKey"`
    Model   string         `json:"model" binding:"required"`
    Year    int            `json:"year" binding:"required"`
    VIN     string         `json:"vin" binding:"required"`
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
    
    // Директория для статических файлов
    router.Static("/static", "./static")
    
    router.LoadHTMLGlob("templates/*")

    router.GET("/", viewCars)
    router.POST("/add", addServiceRecord)       // Добавление записи
    router.POST("/edit", editServiceRecord)     // Редактирование записи
    router.POST("/delete", deleteServiceRecord) // Удаление записи

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

    year, err := strconv.Atoi(yearStr)
    if err != nil {
        c.HTML(http.StatusBadRequest, "index.html", gin.H{"error": "Ошибка в году: " + err.Error()})
        return
    }

    // Найти или создать автомобиль по VIN
    var car Car
    if err := db.Where("vin = ?", vin).First(&car).Error; err != nil {
        car = Car{Model: model, Year: year, VIN: vin}
        db.Create(&car) // Создаем новый автомобиль
    }

    // Создание новой записи об обслуживании
    record := ServiceRecord{
        Date:        c.PostForm("date"),
        Description: c.PostForm("description"),
        Cost:       getCost(c),
        CarID:      car.ID,
    }

    // Создаем новую запись об обслуживании и сохраняем в базе
    db.Create(&record)

    c.Redirect(http.StatusFound, "/")
}

func editServiceRecord(c *gin.Context) {
    recordIDStr := c.PostForm("record_id")
    date := c.PostForm("date")
    description := c.PostForm("description")
    costStr := c.PostForm("cost")

    var record ServiceRecord
    id, err := strconv.ParseUint(recordIDStr, 10, 32)
    if err != nil {
        c.HTML(http.StatusBadRequest, "index.html", gin.H{"error": "Ошибка в ID записи: " + err.Error()})
        return
    }

    if err := db.First(&record, id).Error; err != nil {
        c.HTML(http.StatusInternalServerError, "index.html", gin.H{"error": "Запись не найдена: " + err.Error()})
        return
    }

    record.Date = date
    record.Description = description

    cost, err := strconv.ParseFloat(costStr, 64)
    if err != nil {

        c.HTML(http.StatusBadRequest, "index.html", gin.H{"error": "Ошибка в стоимости: " + err.Error()})
        return
    }
    record.Cost = cost

    if err := db.Save(&record).Error; err != nil {
        c.HTML(http.StatusInternalServerError, "index.html", gin.H{"error": "Ошибка при обновлении записи: " + err.Error()})
        return
    }

    c.Redirect(http.StatusFound, "/")
}

func deleteServiceRecord(c *gin.Context) {
    recordIDStr := c.PostForm("record_id")

    recordID, err := strconv.ParseUint(recordIDStr, 10, 32)
    if err != nil {
        c.HTML(http.StatusBadRequest, "index.html", gin.H{"error": "Ошибка в ID записи: " + err.Error()})
        return
    }

    // Получаем запись обслуживания для доступа к CarID
    var record ServiceRecord
    if err := db.First(&record, recordID).Error; err != nil {
        c.HTML(http.StatusInternalServerError, "index.html", gin.H{"error": "Ошибка при получении записи обслуживания: " + err.Error()})
        return
    }

    // Удаляем запись обслуживания
    if err := db.Delete(&ServiceRecord{}, recordID).Error; err != nil {
        c.HTML(http.StatusInternalServerError, "index.html", gin.H{"error": "Ошибка при удалении записи обслуживания: " + err.Error()})
        return
    }

    // Проверяем, остались ли записи обслуживания для этого автомобиля
    var car Car
    if err = db.Preload("Records").First(&car, record.CarID).Error; err == nil {
        // Если не осталось записей обслуживания, удаляем автомобиль
        if len(car.Records) == 0 {
            if err := db.Delete(&Car{}, car.ID).Error; err != nil {
                c.HTML(http.StatusInternalServerError, "index.html", gin.H{"error": "Ошибка при удалении автомобиля: " + err.Error()})
                return
            }
        }
    }

    c.Redirect(http.StatusFound, "/")
}

// Вспомогательная функция для получения стоимости
func getCost(c *gin.Context) float64 {
    costStr := c.PostForm("cost")
    cost, _ := strconv.ParseFloat(costStr, 64)
    return cost
}

