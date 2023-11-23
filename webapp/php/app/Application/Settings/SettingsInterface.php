<?php

declare(strict_types=1);

namespace App\Application\Settings;

use ErrorException;

interface SettingsInterface
{
    /**
     * @throws ErrorException
     */
    public function get(string $key = ''): mixed;
}
