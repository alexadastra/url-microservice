package storage

import (
	"fmt"
	"log"
	"strconv"
	"sync"
	"url-microservice/config"
	"url-microservice/url_service"
)

type UrlStorage interface {
	Save(url *url_service.Url) error                     // save new url to db
	View() ([]url_service.Url, error)                    // get all urls
	ViewIdByUrl(url string) (int, error)                 // get url info by url string
	ViewUrlByDateAndN(date int, n int) ([]string, error) // get urls with 'n' successful checks starting with 'date'
}

type DataBaseUrlStorage struct {
	mutex sync.RWMutex
}

func NewDataBaseUrlStorage() *DataBaseUrlStorage {
	return &DataBaseUrlStorage{}
}

func (storage *DataBaseUrlStorage) Save(url *url_service.Url) error {
	storage.mutex.Lock()
	defer storage.mutex.Unlock()

	sqlStr := "INSERT INTO urls(url_string, url_method, time_interval, unix_time_added) VALUES ($1, $2, $3, $4)"

	stmt, err := config.DB.Prepare(sqlStr)
	defer stmt.Close()
	if err != nil {
		log.Fatal(err)
	}

	_, err = stmt.Exec(url.Url, url.Method, url.TimeInterval, url.TimeCreated)
	if err != nil {
		return fmt.Errorf("could not save new blog post in memory")
	}
	return nil
}

func (storage *DataBaseUrlStorage) View() ([]url_service.Url, error) {
	rows, err := config.DB.Query("SELECT * FROM urls")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	urls := make([]url_service.Url, 0)
	for rows.Next() {
		url := url_service.Url{}
		err := rows.Scan(&url.Id, &url.Url, &url.Method, &url.TimeInterval, &url.TimeCreated)
		if err != nil {
			return nil, err
		}
		urls = append(urls, url)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return urls, nil
}

func (storage *DataBaseUrlStorage) ViewIdByUrl(url string) (int, error) {
	rows, err := config.DB.Query("SELECT * FROM urls WHERE url_string = '" + url + "'")

	if err != nil {
		return 0, err
	}
	defer rows.Close()

	urls := make([]url_service.Url, 0)
	for rows.Next() {
		url := url_service.Url{}
		err := rows.Scan(&url.Id, &url.Url, &url.Method, &url.TimeInterval, &url.TimeCreated)
		if err != nil {
			return 0, err
		}
		urls = append(urls, url)
	}
	if err = rows.Err(); err != nil {
		return 0, err
	}
	id, err := strconv.ParseInt(urls[0].Id, 10, 64)
	return int(id), nil
}

func (storage *DataBaseUrlStorage) ViewUrlByDateAndN(date int, n int) ([]string, error) {
	rows, err := config.DB.Query("SELECT u.url_string FROM urls u INNER JOIN checks c ON u.url_id = c.url_id " +
		"WHERE + c.unix_time_added >= " + strconv.Itoa(date) + "AND c.status_code >= 200 AND c.status_code < 300" +
		"GROUP BY u.url_string HAVING COUNT(*) >=" + strconv.Itoa(n))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	urls := make([]string, 0)
	for rows.Next() {
		url := url_service.Url{}
		err := rows.Scan(&url.Url)
		if err != nil {
			return nil, err
		}
		urls = append(urls, url.Url)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return urls, nil
}
