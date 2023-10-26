<?php

declare(strict_types=1);

namespace IsuPipe\Stats;

use JsonSerializable;

class LivestreamStatistics implements JsonSerializable
{
    public function __construct(
        public int $rank,
        public int $viewersCount,
        public int $totalReactions,
        public int $totalReports,
        public int $maxTip,
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
            'total_reports' => $this->totalReports,
            'max_tip' => $this->maxTip,
        ];
    }
}
