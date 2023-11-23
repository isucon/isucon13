<?php

declare(strict_types=1);

namespace IsuPipe\User;

use JsonSerializable;

class PostIconResponse implements JsonSerializable
{
    public function __construct(
        public int $id,
    ) {
    }

    /**
     * @return array<string, mixed>
     */
    public function jsonSerialize(): array
    {
        return [
            'id' => $this->id,
        ];
    }
}
