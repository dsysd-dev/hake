package hake

// Hake topic where messages are published
type Topic string

func (a Topic) String() string { return string(a) }

// AccountId for aws which has the correct set of permissions
type AwsAccount string

func (a AwsAccount) String() string { return string(a) }

// aws Region
type AwsRegion string

func (a AwsRegion) String() string { return string(a) }

// accesskey
type AwsAccessKey string

func (a AwsAccessKey) String() string { return string(a) }

// secretkey
type AwsSecretKey string

func (a AwsSecretKey) String() string { return string(a) }
