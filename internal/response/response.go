package response

import "github.com/gofiber/fiber/v2"

type APIResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
	Errors  interface{} `json:"errors"`
}

type ErrorDetail struct {
	Field   string `json:"field,omitempty"`
	Code    string `json:"code"`
	Message string `json:"message"`
}

func Success(
	c *fiber.Ctx,
	status int,
	message string,
	data interface{},
) error {
	return c.Status(status).JSON(APIResponse{
		Success: true,
		Message: message,
		Data:    data,
		Errors:  nil,
	})
}

// Fail sends an error response with the ErrorDetail array you build by hand.
func Fail(
	c *fiber.Ctx,
	status int,
	message string,
	errors []ErrorDetail,
) error {
	return c.Status(status).JSON(APIResponse{
		Success: false,
		Message: message,
		Data:    nil,
		Errors:  errors,
	})
}
