/*
@author: sk
@date: 2024/6/15
*/
package main

type ModelData struct {
	Version        int                 `json:"Version"`
	FileReferences *FileReferencesData `json:"FileReferences"`
	Groups         []*GroupData0       `json:"Groups"`
	HitAreas       []*HitAreaData      `json:"HitAreas"`
}

type HitAreaData struct {
	Id   string `json:"Id"`
	Name string `json:"Name"`
}

type GroupData0 struct {
	Target string   `json:"Target"`
	Name   string   `json:"Name"`
	Ids    []string `json:"Ids"`
}

type FileReferencesData struct {
	Moc         string                    `json:"Moc"`
	Textures    []string                  `json:"Textures"`
	Physics     string                    `json:"Physics"`
	Pose        string                    `json:"Pose"`
	DisplayInfo string                    `json:"DisplayInfo"`
	Expressions []*ExpressionData0        `json:"Expressions"`
	Motions     map[string][]*MotionData0 `json:"Motions"`
	UserData    string                    `json:"UserData"`
}

type ExpressionData0 struct {
	Name string `json:"Name"`
	File string `json:"File"`
}

type MotionData0 struct {
	File        string  `json:"File"`
	FadeInTime  float64 `json:"FadeInTime"`
	FadeOutTime float64 `json:"FadeOutTime"`
	Sound       string  `json:"Sound"`
	MotionSync  string  `json:"MotionSync"`
}

type PhysicData struct {
	Version         int                   `json:"Version"`
	Meta            *MetaData0            `json:"Meta"`
	PhysicsSettings []*PhysicsSettingData `json:"PhysicsSettings"`
}

type PhysicsSettingData struct {
	Id            string             `json:"Id"`
	Input         []*InputData       `json:"Input"`
	Output        []*OutputData      `json:"Output"`
	Vertices      []*VerticesData    `json:"Vertices"`
	Normalization *NormalizationData `json:"Normalization"`
}

type InputData struct {
	Source  *SourceData `json:"Source"`
	Weight  int         `json:"Weight"`
	Type    string      `json:"Type"`
	Reflect bool        `json:"Reflect"`
}

type SourceData struct {
	Target string `json:"Target"`
	Id     string `json:"Id"`
}

type OutputData struct {
	Destination *Vector2 `json:"Destination"`
	VertexIndex int      `json:"VertexIndex"`
	Scale       float64  `json:"Scale"`
	Weight      int      `json:"Weight"`
	Type        string   `json:"Type"`
	Reflect     bool     `json:"Reflect"`
}

type VerticesData struct {
	Position     *Vector2 `json:"Position"`
	Mobility     float64  `json:"Mobility"`
	Delay        float64  `json:"Delay"`
	Acceleration float64  `json:"Acceleration"`
	Radius       int      `json:"Radius"`
}

type NormalizationData struct {
	Position *ValueData `json:"Position"`
	Angle    *ValueData `json:"Angle"`
}

type ValueData struct {
	Minimum float64 `json:"Minimum"`
	Default float64 `json:"Default"`
	Maximum float64 `json:"Maximum"`
}

type MetaData0 struct {
	PhysicsSettingCount int                      `json:"PhysicsSettingCount"`
	TotalInputCount     int                      `json:"TotalInputCount"`
	TotalOutputCount    int                      `json:"TotalOutputCount"`
	VertexCount         int                      `json:"VertexCount"`
	EffectiveForces     *EffectiveForceData      `json:"EffectiveForces"`
	PhysicsDictionary   []*PhysicsDictionaryData `json:"PhysicsDictionary"`
}

type PhysicsDictionaryData struct {
	Id   string `json:"Id"`
	Name string `json:"Name"`
}

type EffectiveForceData struct {
	Gravity *Vector2 `json:"Gravity"`
	Wind    *Vector2 `json:"Wind"`
}

type PoseData struct {
	Type       string          `json:"Type"`
	FadeInTime float64         `json:"FadeInTime"`
	Groups     [][]*GroupData1 `json:"Groups"`
}

type GroupData1 struct {
	Id   string   `json:"Id"`
	Link []string `json:"Link"`
}

type DisplayData struct {
	Version            int               `json:"Version"`
	Parameters         []*ParameterData0 `json:"Parameters"`
	ParameterGroups    []*ParameterData0 `json:"ParameterGroups"`
	Parts              []*PartData       `json:"Parts"`
	CombinedParameters [][]string        `json:"CombinedParameters"`
}

type PartData struct {
	Id   string `json:"Id"`
	Name string `json:"Name"`
}

type ParameterData0 struct {
	Id      string `json:"Id"`
	GroupId string `json:"GroupId"`
	Name    string `json:"Name"`
}

type ExpressionData1 struct {
	Name       string            `json:"-"`
	Type       string            `json:"Type"`
	Parameters []*ParameterData1 `json:"Parameters"`
}

type ParameterData1 struct {
	Id    string  `json:"Id"`
	Value float64 `json:"Value"`
	Blend string  `json:"Blend"`
}

type MotionData1 struct {
	Data    *MotionData0 `json:"-"`
	Version int          `json:"Version"`
	Meta    *MetaData1   `json:"Meta"`
	Curves  []*CurveData `json:"Curves"`
}

type CurveData struct {
	Target      string    `json:"Target"`
	Id          string    `json:"Id"`
	FadeInTime  *float64  `json:"FadeInTime"`
	FadeOutTime *float64  `json:"FadeOutTime"`
	Segments    []float64 `json:"Segments"`
}

type MetaData1 struct {
	Duration             float64 `json:"Duration"`
	Fps                  float64 `json:"Fps"`
	Loop                 bool    `json:"Loop"`
	AreBeziersRestricted bool    `json:"AreBeziersRestricted"`
	CurveCount           int     `json:"CurveCount"`
	TotalSegmentCount    int     `json:"TotalSegmentCount"`
	TotalPointCount      int     `json:"TotalPointCount"`
	UserDataCount        int     `json:"UserDataCount"`
	TotalUserDataSize    int     `json:"TotalUserDataSize"`
}

type UserData0 struct {
	Version  int          `json:"Version"`
	Meta     *MetaData2   `json:"Meta"`
	UserData []*UserData1 `json:"UserData"`
}

type UserData1 struct {
	Target string `json:"Target"`
	Id     string `json:"Id"`
	Value  string `json:"Value"`
}

type MetaData2 struct {
	UserDataCount     int `json:"UserDataCount"`
	TotalUserDataSize int `json:"TotalUserDataSize"`
}
