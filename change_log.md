# gohessian #
本库虽然是个分支，但是我尽力把优化和测试都做好，用于支持dubbogo(github.com/AlexStocks/dubbogo)进行hessian的编解码。

## 说明 ##
---
> 目前gohessian支持hessian protocol 1.2。
> 因为阿里的dubbo支持的是hessian protocol 2.0，所以打算月底前将对gohessian进行改造以支持之。

## develop list ##
---
*本项目其实是从2016/10/21开始进行改造测试的工作了，但2016/10/28才上传到github上，所以dev list只能从29号开始记了*

### 2017-06-13 ###
---
- 1 fix掉decList的bug，做完单测。

### 2017-06-05 ###
---
- 1 重写hessian2 enc/dec后，补充了单测

### 2016-11-04 ###
---
- 1 添加hessian2支持

### 2016-10-28 ###
---
- 1 添加github.com/AlexStocks/gohessian/encode.go:encMapByReflect函数，以编码诸如map[int]string等非map[Any]Any类型的map

### 2016-10-29 ###
---
- 1 修改 github.com/AlexStocks/gohessian/encode.go:encMap & github.com/AlexStocks/gohessian/encode.go:encMapByReflect 两个函数，当map为空的时候防止在buf里面形成垃圾数据

