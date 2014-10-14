goredis
=======

golang redis client, bufferd connection, connection pool

###	Create a new conn?
>		c, e := Dial("127.0.0.1:6379", pwd, CTimeout, RTimeout, WTimeout, alive, *pool)
>		if e != nil {
>			println(e.Error())
>			return
>		}
>
>	If redis do not need AUTH, password ="" 