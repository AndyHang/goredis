goredis
=======

golang redis client, bufferd connection, connection pool, support all redis commands

####	Create a new conn?
>		c, e := Dial("127.0.0.1:6379", pwd, CTimeout, RTimeout, WTimeout, alive, *pool)
>		if e != nil {
>			println(e.Error())
>			return
>		}
>
>	If redis do not need AUTH, password =""

####	A Redis Command.
>		c.GET("mykey")
>		c.SADD("mySets", []string{"a","b","c"})
>		// You can also use this
>		c.Call(CommandName, arg...)

####	Pipeline
>		c.PipeSend("SET", "a", "zyh")
>		c.PipeSend("SET", "b", "zyh")
>		c.PipeSend("SET", "c", "zyh")
>		c.PipeExec()

####	Transaction
>		c.MULTI()
>		c.TransSend("SET", "a", "zyh2")
>		c.TransSend("SET", "b", "zyh3")
>		c.TransExec()


####	Create a new pool?
>		p := NewPool("127.0.0.1:6379", "", maxConnNum, maxIdleSeconds)
>		// get a new conn
>		c := p.Pop()  
>		if c == nil{
>			fmt.Println("get a nil conn")
>		}
>		defer p.Push(c)

####	Create a new multiPool?
>		addresses := []string{"127.0.0.1:6379", "127.0.0.1:9991@1"}
>		mp := NewMultiPool(addresses, maxConnNum, maxIdleSeconds)
>		mp.AddPool("127.0.0.1:9988", maxConnNum+10, maxIdleSeconds+10)
>		addr := "127.0.0.1:6379"
>		c := mp.PopByAddr(addr)
>		mp.PushByAddr(addr, c)
>		key := "myhashes"
>		c = mp.PopByKey(key)
>		mp.PushByKey(key, c)
>	PopByKey and PushByKey based on a Hash Algorithm. You can change your hash Algorithm.
>	The Hash code is implemented in Sum function.
>	AddPool is for a new pool in multiPool that have different maxConnNum and maxIdleSeconds configuration.
>	Note: AddPool must call before use the multiPool, because the map inside multiPool have no mutex.

#### Todo List
+	Consistent Hash
+	Etc...