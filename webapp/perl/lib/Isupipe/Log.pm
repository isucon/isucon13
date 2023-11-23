package Isupipe::Log;
use v5.38;
use utf8;

use Encode ();
use Log::Minimal;

use Exporter 'import';
our @EXPORT = @Log::Minimal::EXPORT;

# warn する時のログフォーマット
#  Log::Minimalと違い、encode してから warnする
$Log::Minimal::PRINT = sub {
    my ( $time, $type, $message, $trace, $raw_message) = @_;
    warn Encode::encode_utf8("$time [$type] $message at $trace\n");
};

# die する時のログフォーマット
#  Log::Minimalと違い、encode してから dieする
$Log::Minimal::DIE = sub {
    my ( $time, $type, $message, $trace, $raw_message) = @_;
    die Encode::encode_utf8("$time [$type] $message at $trace\n");
};

# 色付きの出力設定
#  1: 色あり
#  0: 色なし
$Log::Minimal::COLOR = 1;

# ログを出力するレベル
#  DEBUG: debug 以上のログを出力(LM_DEBUG=1 で出力)
#  INFO: info 以上のログを出力
#  WARN: warn 以上のログを出力
#  ERROR: error 以上のログを出力
#  CRITICAL: critical 以上のログを出力
#  MUTE: ログを出力しない
$Log::Minimal::LOG_LEVEL = 'DEBUG';

# DEBUG のログを出力するか
#  1: 出力する
#  0: 出力しない
$ENV{LM_DEBUG} = ($ENV{PLACK_ENV}||'') ne 'production';

