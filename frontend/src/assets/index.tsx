import BBB from './img/BBB.webp?url';
import hiru from './img/hiru.webp?url';
import hkd from './img/hkd.webp?url';
import isucon11_final from './img/isucon11_final.webp?url';
import isucon12_final from './img/isucon12_final.webp?url';
import isucon12_final_live from './img/isucon12_final_live.webp?url';
import isucon13 from './img/isucon13.webp?url';
import isucon9 from './img/isucon9.webp?url';
import timewarp from './img/timewarp.webp?url';
import ube from './img/ube.webp?url';
import yoru from './img/yoru.webp?url';

export const imageUrls = [
  BBB,
  hiru,
  hkd,
  isucon11_final,
  isucon12_final,
  isucon12_final_live,
  isucon13,
  isucon9,
  timewarp,
  ube,
  yoru,
];

export function getThumbnailUrl(id: number): string {
  return imageUrls[(id - 1) % imageUrls.length];
}
