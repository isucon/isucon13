<?php

declare(strict_types=1);

namespace IsuPipe\Top;

use JsonSerializable;

class Tag implements JsonSerializable
{
    public function __construct(
        public int $id,
        public string $name,
    ) {
    }

    /**
     * @return array<string, mixed>
     */
    public function jsonSerialize(): array
    {
        return [
            'id' => $this->id,
            'name' => $this->name,
        ];
    }
}
