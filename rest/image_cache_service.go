package rest

//@Service
type ImageCacheService struct {
	Bean
	imageCacheDao *ImageCacheDao
}

//初始化方法
func (this *ImageCacheService) Init(context *Context) {

	//手动装填本实例的Bean. 这里必须要用中间变量方可。
	b := context.GetBean(this.imageCacheDao)
	if b, ok := b.(*ImageCacheDao); ok {
		this.imageCacheDao = b
	}

}

//获取某个文件的详情，会把父级依次倒着装进去。如果中途出错，直接抛出异常。
func (this *ImageCacheService) Detail(uuid string) *ImageCache {

	imageCache := this.imageCacheDao.CheckByUuid(uuid)

	return imageCache
}

