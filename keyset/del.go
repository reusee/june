package keyset

type Delete func(
	set Set,
	keys ...Key,
) (
	newSet Set,
	err error,
)

//TODO
