package httplib

type option struct {
	suppressRspLog       bool
	jwtRawData           []byte
	isJwt                bool
	enableJwtEncodedData bool
	cookies              []*cookiePair
	headers              []*headerPair
	retryTimes           int
	suppressLog          bool
	useReadTimeOutClient bool
}

type headerPair struct {
	key   string
	value string
}

type cookiePair struct {
	key   string
	value string
}

type Option func(*option)

func SuppressResponseLog(supress bool) Option {
	return func(o *option) { o.suppressRspLog = supress }
}

func LogJwtEncodedData(enable bool) Option {
	return func(o *option) {
		o.enableJwtEncodedData = enable
	}
}

func JwtRawData(jwtRawData []byte) Option {
	return func(o *option) {
		o.jwtRawData = jwtRawData
		o.isJwt = true
	}
}

func SuppressLog(suppressLog bool) Option {
	return func(o *option) {
		o.suppressLog = suppressLog
	}
}

func ReadTimeOutClient(enableReadTimeOut bool) Option {
	return func(o *option) {
		o.useReadTimeOutClient = enableReadTimeOut
	}
}

func AddHeader(key, value string) Option {
	return func(o *option) {
		o.headers = append(o.headers, &headerPair{key, value})
	}
}

func AddCookie(key, value string) Option {
	return func(o *option) {
		o.cookies = append(o.cookies, &cookiePair{key, value})
	}
}

func RetryTimes(n int) Option {
	return func(o *option) {
		if n > 0 {
			o.retryTimes = n
		}
	}
}
