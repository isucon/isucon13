<?php

declare(strict_types=1);

namespace App\Application\Settings;

use ErrorException;

class Settings implements SettingsInterface
{
    /**
     * @param array<string, mixed> $settings
     */
    public function __construct(private array $settings)
    {
    }

    /**
     * @throws ErrorException
     */
    public function get(string $key = ''): mixed
    {
        if (empty($key)) {
            return $this->settings;
        }

        return $this->settings[$key] ?? throw new ErrorException('undefined setting key');
    }
}
