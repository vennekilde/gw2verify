package server

import (
	"context"
	"io"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/vennekilde/gw2verify/v2/internal/api"
	"github.com/vennekilde/gw2verify/v2/internal/orm"
)

// (GET /v1/services/{service_uuid}/properties)
func (e *Endpoints) GetServiceProperties(c *gin.Context, serviceUuid api.ServiceUuid) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var properties []api.Property
	err := orm.DB().NewSelect().
		Model(&properties).
		Where("service_uuid = ?", serviceUuid).
		Scan(ctx)
	if err != nil {
		ThrowReqError(c, err.Error(), nil, http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, &properties)
}

// (GET /v1/services/{service_uuid}/properties/{subject})
func (e *Endpoints) GetServiceSubjectProperties(c *gin.Context, serviceUuid api.ServiceUuid, subject api.Subject) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var properties []api.Property
	err := orm.DB().NewSelect().
		Model(&properties).
		Where("service_uuid = ? AND subject LIKE ?", serviceUuid, subject).
		Scan(ctx)
	if err != nil {
		ThrowReqError(c, err.Error(), nil, http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, &properties)
}

// (PUT /v1/services/{service_uuid}/properties/{subject})
func (e *Endpoints) PutServiceSubjectProperties(c *gin.Context, serviceUuid api.ServiceUuid, subject api.Subject) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var properties []api.Property
	// decode request
	err := c.Bind(&properties)
	if err != nil {
		ThrowReqError(c, err.Error(), err, http.StatusBadRequest)
		return
	}

	_, err = orm.DB().NewInsert().
		Model(&properties).
		Value("db_updated", "NOW()").
		Value("service_uuid", "?", serviceUuid).
		Value("subject", "?", subject).
		On(`CONFLICT ("service_uuid", "subject", "name") DO UPDATE`).
		Set("name = EXCLUDED.name, value = EXCLUDED.value, db_updated = EXCLUDED.db_updated").
		Exec(ctx)
	if err != nil {
		ThrowReqError(c, err.Error(), nil, http.StatusInternalServerError)
		return
	}

	c.Status(http.StatusOK)
}

// (GET /v1/services/{service_uuid}/properties/{subject}/{property_name})
func (e *Endpoints) GetServiceSubjectProperty(c *gin.Context, serviceUuid api.ServiceUuid, subject api.Subject, propertyName api.PropertyName) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var property api.Property
	err := orm.DB().NewSelect().
		Model(&property).
		Where("service_uuid = ? AND subject = ? AND name = ?", serviceUuid, subject, propertyName).
		Scan(ctx)
	if err != nil {
		ThrowReqError(c, err.Error(), nil, http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, &property)
}

// (PUT /v1/services/{service_uuid}/properties/{subject}/{property_name})
func (e *Endpoints) PutServiceSubjectProperty(c *gin.Context, serviceUuid api.ServiceUuid, subject api.Subject, propertyName api.PropertyName) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// decode request
	value, err := io.ReadAll(c.Request.Body)
	if err != nil {
		ThrowReqError(c, err.Error(), err, http.StatusBadRequest)
		return
	}

	property := api.Property{
		Name:  propertyName,
		Value: string(value),
	}
	_, err = orm.DB().NewInsert().
		Model(&property).
		Value("db_updated", "NOW()").
		Value("service_uuid", "?", serviceUuid).
		Value("subject", "?", subject).
		On(`CONFLICT ("service_uuid", "subject", "name") DO UPDATE`).
		Set("name = EXCLUDED.name, value = EXCLUDED.value, db_updated = EXCLUDED.db_updated").
		Exec(ctx)
	if err != nil {
		ThrowReqError(c, err.Error(), nil, http.StatusInternalServerError)
		return
	}

	c.Status(http.StatusOK)
}
