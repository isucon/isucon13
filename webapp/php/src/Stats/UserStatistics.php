<?php

declare(strict_types=1);

namespace IsuPipe\Stats;

use JsonSerializable;

class UserStatistics implements JsonSerializable
{
    public function __construct(
        public int $rank,
        public int $viewersCount,
        public int $totalReactions,
        public int $totalLivecomments,
        public int $totalTip,
        public string $favoriteEmoji,
    ) {
    }

    /**
     * @return array<string, mixed>
     */
    public function jsonSerialize(): array
    {
        return [
            'rank' => $this->rank,
            'viewers_count' => $this->viewersCount,
            'total_reactions' => $this->totalReactions,
            'total_livecomments' => $this->totalLivecomments,
            'total_tip' => $this->totalTip,
            'favorite_emoji' => $this->favoriteEmoji,
        ];
    }
}
