package metrics

func (w *Watcher) GetMemUsage() *int64 {
	if w == nil {
		return nil
	}

	w.mu.Lock()
	defer w.mu.Unlock()

	memUsage := w.data.MemUsage

	return memUsage
}
