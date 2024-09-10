package handler

import (
	"errors"
	"fmt"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgconn"
	"go.uber.org/zap"
	"key-server/database"
	"key-server/logger"
	"key-server/model"
	"key-server/util"
	"net/http"
	"os"
	"reflect"
	"slices"
	"strings"
)

// Define the allowed quality values
var allowedQualityValues = map[string]bool{
	"HD":    true,
	"SD":    true,
	"UHD1":  true,
	"UHD2":  true,
	"AUDIO": true,
}

type KeyData struct {
	KeyID string `json:"keyId"`
	KeyIV string `json:"keyIv"`
	Key   string `json:"key"`
}

type KeyGenRequest struct {
	ContentID string `json:"contentId" validate:"required"`
	PackageID string `json:"packageId" validate:"required"`
	// Quality    string `json:"quality" validate:"required,oneof=ALL AUDIO HD SD"`
	Quality    []string `json:"quality" validate:"required,validateQuality"`
	ProviderID string   `json:"providerId" validate:"required"`
	DrmScheme  []string `json:"drmScheme" validate:"required,validateDrmScheme"`
}

type KeyGenResponse struct {
	ContentID  string               `json:"contentId"`
	PackageID  string               `json:"packageId"`
	ProviderID string               `json:"providerId"`
	DrmScheme  []string             `json:"drmScheme"`
	Keys       []map[string]KeyData `json:"keys"`
}

type KeyGetResponse struct {
	ContentID  string               `json:"contentId"`
	PackageID  string               `json:"packageId"`
	ProviderID string               `json:"providerId"`
	Keys       []map[string]KeyData `json:"keys"`
}

func validateDrmScheme(fl validator.FieldLevel) bool {
	// Get the value of the field
	value := fl.Field()

	// Check if the value is an array or slice
	if value.Kind() != reflect.Slice && value.Kind() != reflect.Array {
		return false
	}

	// Get the value of the DrmScheme field from the parent struct
	drmSchemeValue := fl.Parent().FieldByName("DrmScheme").Interface().([]string)

	// Validate each element in the DrmScheme array
	for _, drmScheme := range drmSchemeValue {
		if drmScheme != "WV" && drmScheme != "FP" && drmScheme != "PR" {
			return false
		}
	}

	// All elements are valid
	return true
}

// Custom validator function for the Quality field
func validateQuality(fl validator.FieldLevel) bool {
	// Get the value of the field
	value := fl.Field()

	// Check if the value is an array or slice
	if value.Kind() != reflect.Slice && value.Kind() != reflect.Array {
		return false
	}

	// Iterate over each element in the array/slice
	for i := 0; i < value.Len(); i++ {
		// Get the value of the element
		elem := fmt.Sprintf("%v", value.Index(i).Interface())

		// Check if the element is in the allowed quality values
		if !allowedQualityValues[elem] {
			return false
		}
	}

	// All elements are valid
	return true
}

func GenerateKey(ctx *fiber.Ctx) error {

	var keyGenInput KeyGenRequest

	if err := ctx.BodyParser(&keyGenInput); err != nil {
		return ctx.Status(http.StatusBadRequest).
			JSON(util.ErrorResponse("Invalid request body", err, 5001))
	}

	validate := validator.New()
	// Register the custom validator function
	if err := validate.RegisterValidation("validateQuality", validateQuality); err != nil {
		fmt.Println("Failed to register custom validator:", err)
		return ctx.Status(fiber.StatusBadRequest).JSON(
			util.ErrorResponse("Invalid request body", err, 5020))
	}

	if err := validate.RegisterValidation("validateDrmScheme", validateDrmScheme); err != nil {
		fmt.Println("Failed to register custom validator:", err)
		return ctx.Status(fiber.StatusBadRequest).JSON(
			util.ErrorResponse("Invalid request body", err, 5020))
	}

	if err := validate.Struct(keyGenInput); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(
			util.ErrorResponse("Invalid request body", err, 5002))
	}
	db := database.DB.Db
	keys := make([]map[string]KeyData, 0)

	var keysSlice []model.EncryptionKey

	qualities := keyGenInput.Quality
	containsFP := slices.Contains(keyGenInput.DrmScheme, "FP") && len(keyGenInput.DrmScheme) > 1
	for _, drmScheme := range keyGenInput.DrmScheme {
		if containsFP {
			keyIV := util.Generate16ByteHex()
			key := util.Generate16ByteHex()

			for _, quality := range qualities {
				// Generate 16-byte hex strings for key ID, IV, and key
				keyID := util.Generate16ByteHex()
				keyIV := keyIV
				key := key
				keyData := KeyData{
					KeyID: keyID,
					KeyIV: keyIV,
					Key:   key,
				}
				keysModel := model.EncryptionKey{
					ContentID:  keyGenInput.ContentID,
					PackageID:  keyGenInput.PackageID,
					Quality:    quality,
					ProviderID: keyGenInput.ProviderID,
					DrmScheme:  strings.Join(keyGenInput.DrmScheme, ","),
					KeyID:      keyID,
					KeyIV:      keyIV,
					Key:        key,
				}
				keysSlice = append(keysSlice, keysModel)

				keys = append(keys, map[string]KeyData{quality: keyData})
			}
			break
		} else {
			if drmScheme == "WV" || drmScheme == "PR" {
				//qualities := []string{"HD", "SD", "AUDIO"}
				for _, quality := range qualities {
					// Generate 16-byte hex strings for key ID, IV, and key
					keyID := util.Generate16ByteHex()
					keyIV := util.Generate16ByteHex()
					key := util.Generate16ByteHex()
					keyData := KeyData{
						KeyID: keyID,
						KeyIV: keyIV,
						Key:   key,
					}
					keysModel := model.EncryptionKey{
						ContentID:  keyGenInput.ContentID,
						PackageID:  keyGenInput.PackageID,
						Quality:    quality,
						ProviderID: keyGenInput.ProviderID,
						DrmScheme:  strings.Join(keyGenInput.DrmScheme, ","),
						KeyID:      keyID,
						KeyIV:      keyIV,
						Key:        key,
					}
					keysSlice = append(keysSlice, keysModel)

					keys = append(keys, map[string]KeyData{quality: keyData})
				}
				break
			} else if len(keyGenInput.DrmScheme) == 1 && drmScheme == "FP" {
				qualities := []string{"HD", "SD", "AUDIO"}

				keyIV := util.Generate16ByteHex()
				key := util.Generate16ByteHex()

				for _, quality := range qualities {

					keyID := util.Generate16ByteHex()
					keyIV := keyIV
					key := key
					keyData := KeyData{
						KeyID: keyID,
						KeyIV: keyIV,
						Key:   key,
					}
					keysModel := model.EncryptionKey{
						ContentID:  keyGenInput.ContentID,
						PackageID:  keyGenInput.PackageID,
						Quality:    quality,
						ProviderID: keyGenInput.ProviderID,
						DrmScheme:  strings.Join(keyGenInput.DrmScheme, ","),
						KeyID:      keyID,
						KeyIV:      keyIV,
						Key:        key,
					}
					keysSlice = append(keysSlice, keysModel)

					keys = append(keys, map[string]KeyData{quality: keyData})
				}
			}
		}
	}

	result := db.Create(&keysSlice)
	if result.Error != nil {
		log := logger.GetForFile("db-errors")
		log.Error("Error while inserting keys to DB", zap.Error(result.Error))

		return ctx.Status(http.StatusInternalServerError).
			JSON(util.ErrorResponse("Could not generate keys", result.Error, 5003))
	}

	responseData := KeyGenResponse{
		ContentID:  keyGenInput.ContentID,
		PackageID:  keyGenInput.PackageID,
		ProviderID: keyGenInput.ProviderID,
		DrmScheme:  keyGenInput.DrmScheme,
		Keys:       keys,
	}

	return ctx.Status(fiber.StatusCreated).JSON(util.SuccessResponse(responseData, "Key generated"))

}

func GetKey(ctx *fiber.Ctx) error {

	contentID := ctx.Params("contentId")
	packageID := ctx.Params("packageId")

	// Parse the request body to extract drmScheme and quality parameters
	var requestBody struct {
		DrmScheme []string `json:"drmScheme"`
		Quality   []string `json:"quality" validate:"required"`
	}

	if err := ctx.BodyParser(&requestBody); err != nil {
		return ctx.Status(http.StatusBadRequest).
			JSON(util.ErrorResponse("Invalid request body", err, 5001))
	}

	validate := validator.New()
	if err := validate.Struct(requestBody); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(
			util.ErrorResponse("Invalid request body", err, 5002))
	}

	// Check if the request body is empty
	/*if len(ctx.Request().Body()) == 0 {
		// If the request body is empty, set default values for DrmScheme and Quality
		requestBody.DrmScheme = []string{} // or any other default value
		requestBody.Quality = []string{}   // or any other default value
	} else {
		// Parse the request body
		if err := ctx.BodyParser(&requestBody); err != nil {
			return ctx.Status(http.StatusBadRequest).
				JSON(util.ErrorResponse("Invalid request body", err, 5001))
		}
	}*/

	db := database.DB.Db

	// Build the query dynamically based on the presence of parameters and path variables
	query := db.Where("content_id = ? AND package_id = ?", contentID, packageID)

	if len(requestBody.DrmScheme) > 0 {
		// Use PostgreSQL's ANY operator to check if any of the values match
		subQuery := db.Where("? = ANY(string_to_array(drm_scheme, ','))", requestBody.DrmScheme[0])
		for i := 1; i < len(requestBody.DrmScheme); i++ {
			subQuery = subQuery.Or("? = ANY(string_to_array(drm_scheme, ','))", requestBody.DrmScheme[i])
		}
		query = query.Where(subQuery)
	}

	if len(requestBody.Quality) > 0 {
		query = query.Where("quality IN (?)", requestBody.Quality)
	}

	// Execute the query
	var keys []model.EncryptionKey
	if err := query.Find(&keys).Error; err != nil {
		log := logger.GetForFile("db-errors")
		log.Error("Error while querying DB", zap.Error(err))

		return ctx.Status(http.StatusInternalServerError).
			JSON(util.ErrorResponse("Could not get keys", err, 5003))
	}

	if len(keys) == 0 {
		return ctx.Status(http.StatusNotFound).
			JSON(util.ErrorResponse("No key found", errors.New("no data found"), 5004))
	}

	// Construct the response object
	response := KeyGetResponse{
		ContentID:  contentID,
		PackageID:  packageID,
		ProviderID: keys[0].ProviderID, // Assume all keys have the same ProviderID
		Keys:       make([]map[string]KeyData, len(keys)),
	}

	// Populate the Keys array
	for i, key := range keys {
		keyData := KeyData{
			KeyID: key.KeyID,
			KeyIV: key.KeyIV,
			Key:   key.Key,
		}
		response.Keys[i] = map[string]KeyData{
			key.Quality: keyData,
		}
	}

	return ctx.Status(fiber.StatusOK).JSON(util.SuccessResponse(response, "Keys returned"))

}

func GenerateStaticKey(ctx *fiber.Ctx) error {

	var keyGenInput KeyGenRequest

	if err := ctx.BodyParser(&keyGenInput); err != nil {
		return ctx.Status(http.StatusBadRequest).
			JSON(util.ErrorResponse("Invalid request body", err, 5001))
	}

	validate := validator.New()
	// Register the custom validator function
	if err := validate.RegisterValidation("validateQuality", validateQuality); err != nil {
		fmt.Println("Failed to register custom validator:", err)
		return ctx.Status(fiber.StatusBadRequest).JSON(
			util.ErrorResponse("Invalid request body", err, 5020))
	}

	if err := validate.RegisterValidation("validateDrmScheme", validateDrmScheme); err != nil {
		fmt.Println("Failed to register custom validator:", err)
		return ctx.Status(fiber.StatusBadRequest).JSON(
			util.ErrorResponse("Invalid request body", err, 5020))
	}

	if err := validate.Struct(keyGenInput); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(
			util.ErrorResponse("Invalid request body", err, 5002))
	}
	db := database.DB.Db

	// Static keys to be returned
	staticKeys := []map[string]KeyData{
		{
			"AUDIO": {
				KeyID: os.Getenv("AUDIO_KEY_ID"),
				KeyIV: os.Getenv("AUDIO_KEY_IV"),
				Key:   os.Getenv("AUDIO_KEY"),
			},
		},
		{
			"HD": {
				KeyID: os.Getenv("HD_KEY_ID"),
				KeyIV: os.Getenv("HD_KEY_IV"),
				Key:   os.Getenv("HD_KEY"),
			},
		},
		{
			"SD": {
				KeyID: os.Getenv("SD_KEY_ID"),
				KeyIV: os.Getenv("SD_KEY_IV"),
				Key:   os.Getenv("SD_KEY"),
			},
		},
	}

	keysSlice := []model.EncryptionKey{
		{
			ContentID:  keyGenInput.ContentID,
			PackageID:  keyGenInput.PackageID,
			Quality:    "AUDIO",
			ProviderID: keyGenInput.ProviderID,
			DrmScheme:  strings.Join(keyGenInput.DrmScheme, ","),
			KeyID:      os.Getenv("AUDIO_KEY_ID"),
			KeyIV:      os.Getenv("AUDIO_KEY_IV"),
			Key:        os.Getenv("AUDIO_KEY"),
		},
		{
			ContentID:  keyGenInput.ContentID,
			PackageID:  keyGenInput.PackageID,
			Quality:    "HD",
			ProviderID: keyGenInput.ProviderID,
			DrmScheme:  strings.Join(keyGenInput.DrmScheme, ","),
			KeyID:      os.Getenv("HD_KEY_ID"),
			KeyIV:      os.Getenv("HD_KEY_IV"),
			Key:        os.Getenv("HD_KEY"),
		},
		{
			ContentID:  keyGenInput.ContentID,
			PackageID:  keyGenInput.PackageID,
			Quality:    "SD",
			ProviderID: keyGenInput.ProviderID,
			DrmScheme:  strings.Join(keyGenInput.DrmScheme, ","),
			KeyID:      os.Getenv("SD_KEY_ID"),
			KeyIV:      os.Getenv("SD_KEY_IV"),
			Key:        os.Getenv("SD_KEY"),
		},
	}

	result := db.Create(&keysSlice)

	if result.Error != nil {
		var pgError *pgconn.PgError
		if errors.As(result.Error, &pgError) && pgError.Code == "23505" {
			// fmt.Println("Duplicate key error: ", pgError.Detail)
			responseData := KeyGenResponse{
				ContentID:  keyGenInput.ContentID,
				PackageID:  keyGenInput.PackageID,
				ProviderID: keyGenInput.ProviderID,
				DrmScheme:  keyGenInput.DrmScheme,
				Keys:       staticKeys,
			}
			return ctx.Status(fiber.StatusCreated).JSON(util.SuccessResponse(responseData, "Key generated"))
		} else {
			log := logger.GetForFile("db-errors")
			log.Error("Error while inserting keys to DB", zap.Error(result.Error))
			return ctx.Status(http.StatusInternalServerError).
				JSON(util.ErrorResponse("Could not generate keys", result.Error, 5003))
		}

	}

	responseData := KeyGenResponse{
		ContentID:  keyGenInput.ContentID,
		PackageID:  keyGenInput.PackageID,
		ProviderID: keyGenInput.ProviderID,
		DrmScheme:  keyGenInput.DrmScheme,
		Keys:       staticKeys,
	}

	return ctx.Status(fiber.StatusCreated).JSON(util.SuccessResponse(responseData, "Key generated"))

}
