include "./user/user.thrift"
include "./interviews/interviews.thrift"
include "./mianshi/mianshi.thrift"
include "./prediction/prediction.thrift"

namespace go interview


service UserService extends user.UserService {}
service InterviewsService extends interviews.InterviewsService {}
service MianshiService extends mianshi.MianshiService {}
service PredictionService extends prediction.PredictionService {}