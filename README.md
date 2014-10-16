goredis
=======

golang redis client, bufferd connection, connection pool, support all redis commands，
欢迎大家批评指正，更欢迎大家加入进来。

####	Create a new conn?
>		c, e := Dial("127.0.0.1:6379", pwd, CTimeout, RTimeout, WTimeout, alive, *pool)
>		if e != nil {
>			println(e.Error())
>			return
>		}
>
>	如果redis不需要AUTH认证, password =""

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
>	PopByKey和PushByKey是对参数key进行hash，然后选出固定的redis。你可以使用自己的hash算法，具体实现在Sum函数中。
>	AddPool函数，会在multiPool中新加入一个Pool，maxConnNum和maxIdleSeconds可以和初始化multiPool的时候不同，
>	注意：必须在init的时候按顺序添加不同参数的Pool，因为multiPool里面没有对pool slice加锁

####	Use Lua script?
###### Lua 脚本是针对pool结构的，每个pool有一个script map，使用者可以预先编写好需要用到的脚本，通过script load函数生成sha1
###### 然后，放入map中，下次方便调用，且不用每次都编译，而且可能节省带宽，变相的增加了访问的速度

###### 需要注意的一点是，如果返回lua的table，所以只能是整数，且从1开始顺序的。
###### 具体关于redis lua的说明请参考<a href="http://redis.io/commands/eval">http://redis.io/commands/eval</a>
>		p := NewPool("10.16.15.121:9731", "", 10, 10)
>		c := p.Pop()
>		if c == nil {
>			return
>		}

>		scriptA := `
>		local ttl = redis.call("ttl",KEYS[1])
>		local key = redis.call("get",KEYS[1])
>
>		local rTable = {}
>		rTable[1] = ttl
>		rTable[2] = key
>		return rTable
>		`
>		sha1, e := c.SCRIPTLOAD(scriptA)
>		if e != nil {
>			fmt.Println("script load error = ", e.Error())
>			return
>		}


#### Todo List
+	Consistent Hash
+	Etc...