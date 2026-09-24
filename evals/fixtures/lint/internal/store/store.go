package store

type Item struct{ ID string }

func NewItem(id string) Item { return Item{ID: id} }
