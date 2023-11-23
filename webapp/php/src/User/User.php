<?php

declare(strict_types=1);

namespace IsuPipe\User;

use JsonSerializable;

class User implements JsonSerializable
{
    public function __construct(
        public int $id,
        public string $name,
        public ?string $displayName = null,
        public ?string $description = null,
        public ?Theme $theme = null,
        public ?string $iconHash = null,
    ) {
    }

    /**
     * @return array<string, mixed>
     */
    public function jsonSerialize(): array
    {
        $data = [
            'id' => $this->id,
            'name' => $this->name,
        ];

        if (!is_null($this->displayName)) {
            $data['display_name'] = $this->displayName;
        }

        if (!is_null($this->description)) {
            $data['description'] = $this->description;
        }

        if (!is_null($this->theme)) {
            $data['theme'] = $this->theme;
        }

        if (!is_null($this->iconHash)) {
            $data['icon_hash'] = $this->iconHash;
        }

        return $data;
    }
}
