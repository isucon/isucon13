package config

// NOTE: ベンチマーカー実行ログを当該パスに書き出す
//
//	supervisorがそれを拾い、ポータルにPOSTする
var StaffLogPath string = "/tmp/result.json"
var ContestantLogPath string = "/tmp/staff.log"
var ResultPath string = "/tmp/contestant.log"
var FinalcheckPath string = "/tmp/finalcheck.json"
