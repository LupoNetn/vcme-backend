package call

import (
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
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
	type callParams struct {
		Title       string      `json:"title"`
		Description pgtype.Text `json:"description"`
		HostID      uuid.UUID   `json:"host_id"`
	}
	var call callParams
	if err := c.ShouldBindJSON(&call); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	link := util.GenerateCallLink("jumbotronm", call.Title)

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

func (h *Handler) ListCallsByHostID(c *gin.Context) {
	hostIDStr := c.Param("host_id")
	hostID, err := uuid.Parse(hostIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid host id"})
		return
	}

	calls, err := h.service.ListAllCallsByID(c.Request.Context(), hostID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "calls successfully retrieved",
		"calls":   calls,
	})
}

func (h *Handler) ListAllCalls(c *gin.Context) {
	calls, err := h.service.ListAllCalls(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "all calls successfully retrieved",
		"calls":   calls,
	})
}
func (h *Handler) GetCallByLink(c *gin.Context) {
	link := c.Param("link")
	// Gin wildcard params include the leading slash, so we trim it
	link = strings.TrimPrefix(link, "/")
	log.Printf("Searching for call link: [%s]", link)

	call, err := h.service.GetCallByLink(c.Request.Context(), link)
	if err != nil {
		log.Printf("Call not found for link: [%s], error: %v", link, err)
		c.JSON(http.StatusNotFound, gin.H{"error": "call not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "call successfully retrieved",
		"call":    call,
	})
}

func (h *Handler) EndCall(c *gin.Context) {
	type endCallParams struct {
		UserID           uuid.UUID   `json:"user_id"`
		CallID           uuid.UUID   `json:"call_id"`
		Participant      pgtype.Text `json:"participant"`
		Type             pgtype.Text `json:"type"`
		Time             pgtype.Text `json:"time"`
		CallTitle        pgtype.Text `json:"call_title"`
		Duration         string      `json:"duration"`
		ParticipantCount int32       `json:"participant_count"`
	}
	var endCall endCallParams
	log.Printf("user foreign id = %v", endCall.UserID)
	if err := c.ShouldBindJSON(&endCall); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	t := pgtype.Timestamptz{
		Time:  time.Now(),
		Valid: true,
	}

	endP := db.EndCallParams{
		UserID:           endCall.UserID,
		CallID:           endCall.CallID,
		Participant:      endCall.Participant,
		Type:             endCall.Type,
		Time:             endCall.Time,
		CallTitle:        endCall.CallTitle,
		Duration:         endCall.Duration,
		ParticipantCount: endCall.ParticipantCount,
		CreatedAt:        t,
	}

	endedCall, err := h.service.EndCall(c.Request.Context(), endP)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}


	c.JSON(http.StatusOK, gin.H{
		"message": "call successfully ended",
		"call":    endedCall,
	})
}

func (h *Handler) GetCallLogsByUserID(c *gin.Context) {
	userIDStr := c.Param("user_id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}

	callLogs, err := h.service.GetCallLogsByUserID(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":  "call logs successfully retrieved",
		"call_logs": callLogs,
	})
}
