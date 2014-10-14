goredis
=======

golang redis client, bufferd connection, connection pool

###	Create a new conn?
>		c, e := Dial("127.0.0.1:6379", password, ConnectTimeout, ReadTimeout, WriteTimeout, keepAlive, *pool)
>		if e != nil {
>			println(e.Error())
>			return
>		}
>
>	If redis do not need AUTH, password ="" 