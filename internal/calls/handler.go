package call

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/luponetn/vcme/internal/db"
	"github.com/luponetn/vcme/internal/util"
)

type Handler struct {
	service Service
}

func NewHandler(s Service) *Handler {
	return &Handler{
		service: s,
	}
}

// call handlers implementations
func (h *Handler) CreateCallLink(c *gin.Context) {
	var call CreateCallRequestParams
	if err := c.ShouldBindJSON(&call); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	link := util.GenerateCallLink(call.Title, "vcme")

	params := db.CreateCallLinkParams{
		Title:       call.Title,
		Description: call.Description,
		CallLink:    link,
		HostID:      call.HostID,
	}

	createdCall, err := h.service.CreateCallLink(c.Request.Context(), params)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":  "call successfully created",
		"callLink": createdCall.CallLink,
		"call":     createdCall,
	})
}
