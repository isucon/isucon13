<?php

declare(strict_types=1);

use App\Application\Settings\Settings;
use App\Application\Settings\SettingsInterface;
use DI\ContainerBuilder;
use Monolog\Logger;

return function (ContainerBuilder $containerBuilder) {

    // Global Settings Object
    $containerBuilder->addDefinitions([
        SettingsInterface::class => function () {
            return new Settings([
                'displayErrorDetails' => true, // Should be set to false in production
                'logError'            => true,
                'logErrorDetails'     => true,
                'logger' => [
                    'name' => 'webapp',
                    'path' => 'php://stdout',
                    'level' => Logger::DEBUG,
                ],
                'database' => [
                    // 環境変数がセットされていなかった場合でも一旦動かせるように、デフォルト値を入れておく
                    // この挙動を変更して、エラーを出すようにしてもいいかもしれない
                    'host' => getenv('ISUCON13_MYSQL_DIALCONFIG_ADDRESS') ?: '127.0.0.1',
                    'port' => getenv('ISUCON13_MYSQL_DIALCONFIG_PORT') ?: '3306',
                    'database' => getenv('ISUCON13_MYSQL_DIALCONFIG_DATABASE') ?: 'isupipe',
                    'username' => getenv('ISUCON13_MYSQL_DIALCONFIG_USER') ?: 'isucon',
                    'password' => getenv('ISUCON13_MYSQL_DIALCONFIG_PASSWORD') ?: 'isucon',
                ],
                'powerdns' => [
                    'subdomainAddress' => getenv('ISUCON13_POWERDNS_SUBDOMAIN_ADDRESS') ?:
                        throw new ErrorException('ISUCON13_POWERDNS_SUBDOMAIN_ADDRESS must be provided'),
                ],
            ]);
        },
    ]);
};
