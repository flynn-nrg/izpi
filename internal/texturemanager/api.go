package texturemanager

type TextureManager interface {
	StoreTextureInCache()
	GetTextureFromCache()
	RemoveTextureFromCache()
}
