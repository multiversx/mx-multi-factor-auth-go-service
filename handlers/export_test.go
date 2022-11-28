package handlers

// RemoveBucket -
func (ih *indexHandler) RemoveBucket(bucketID uint32) {
	delete(ih.indexBuckets, bucketID)
}
