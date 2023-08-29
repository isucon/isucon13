package main

import "github.com/labstack/echo/v4"

// ユーザ登録API
// POST /user
func userRegisterHandler(c echo.Context) error {
	return nil
}

// ユーザログインAPI
// POST /login
func loginHandler(c echo.Context) error {
	return nil
}

// ユーザ詳細API
// GET /user/:userid
func userHandler(c echo.Context) error {
	return nil
}

// ユーザ
// XXX セッション情報返すみたいな？
// GET /user
func userSessionHandler(c echo.Context) error {
	return nil
}

// ユーザが登録しているチャンネル一覧
// GET /user/:user_id/channel
func userChannelHandler(c echo.Context) error {
	return nil
}

// チャンネル登録
// POST /user/:user_id/channel/:channelid/subscribe
func subscribeChannelHandler(c echo.Context) error {
	return nil
}

// チャンネル登録解除
// POST /user/:user_id/channel/:channelid/unsubscribe
func unsubscribeChannelHandler(c echo.Context) error {
	return nil
}

// チャンネル情報
// GET /channel/:channel_id
func channelHandler(c echo.Context) error {
	return nil
}

// チャンネル登録者数
// GET /channel/:channel_id/subscribers
func channelSubscribersHandler(c echo.Context) error {
	return nil
}

// チャンネルの動画一覧
// GET /channel/:channel_id/movie
func channelMovieHandler(c echo.Context) error {
	return nil
}

// チャンネル作成
// POST /channel
func createChannelHandler(c echo.Context) error {
	return nil
}

// チャンネル編集
// PUT /channel/:channel_id
func updateChannelHandler(c echo.Context) error {
	return nil
}

// チャンネル削除
// DELETE /channel/:channel_id
func deleteChannelHandler(c echo.Context) error {
	return nil
}
