package handlers

import (
	"context"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgconn"
)

// StatusClientClosedRequest is the conventional (nginx) status for a request
// the client abandoned before the response was written. It keeps abandoned
// reads out of the 5xx error budget and out of error-level request logs.
const StatusClientClosedRequest = 499

const pgStatementTimeoutCode = "57014"

// writeReadError maps a failed read-only query to a response. A client that
// went away gets 499 with no body, a query that outlived its budget gets 504,
// and everything else keeps the handler's existing 500 message. The error is
// attached to the context so the request logger still records it.
func writeReadError(c *gin.Context, err error, message string) {
	requestCtx := c.Request.Context()
	switch {
	case errors.Is(err, context.Canceled), errors.Is(requestCtx.Err(), context.Canceled):
		c.AbortWithStatus(StatusClientClosedRequest)
	case errors.Is(err, context.DeadlineExceeded), errors.Is(requestCtx.Err(), context.DeadlineExceeded), isStatementTimeout(err):
		_ = c.Error(err)
		c.AbortWithStatusJSON(http.StatusGatewayTimeout, gin.H{"error": message + ": timed out"})
	default:
		_ = c.Error(err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": message})
	}
}

func isStatementTimeout(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == pgStatementTimeoutCode
}
