package middleware

// GetBasePath -
func (mm *metricsMiddleware) GetBasePath(path string) string {
	return mm.getBasePath(path)
}
