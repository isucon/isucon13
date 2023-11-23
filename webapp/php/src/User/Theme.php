<?php

declare(strict_types=1);

namespace IsuPipe\User;

use JsonSerializable;

class Theme implements JsonSerializable
{
    public function __construct(
        public int $id,
        public bool $darkMode,
    ) {
    }

    /**
     * @return array<string, mixed>
     */
    public function jsonSerialize(): array
    {
        return [
            'id' => $this->id,
            'dark_mode' => $this->darkMode,
        ];
    }
}
