import type { Variants } from 'framer-motion'

const CARD_ANIMATION_THRESHOLD = 18

export function shouldAnimateSkillCards(itemCount: number) {
  return itemCount <= CARD_ANIMATION_THRESHOLD
}

export const pageVariants: Variants = {
  initial: { opacity: 0 },
  animate: { opacity: 1, transition: { duration: 0.16, ease: 'easeOut' } },
  exit: { opacity: 0, transition: { duration: 0.1 } },
}

export const gridContainerVariants = (itemCount: number): Variants => ({
  animate: {
    transition: {
      staggerChildren: shouldAnimateSkillCards(itemCount) ? 0.025 : 0,
    },
  },
})

export const cardVariants: Variants = {
  initial: { opacity: 0, y: 12, scale: 0.97 },
  animate: {
    opacity: 1,
    y: 0,
    scale: 1,
    transition: { duration: 0.2, ease: 'easeOut' },
  },
}

export const dialogVariants: Variants = {
  initial: { opacity: 0, scale: 0.94 },
  animate: {
    opacity: 1,
    scale: 1,
    transition: { duration: 0.2, ease: [0.16, 1, 0.3, 1] },
  },
  exit: {
    opacity: 0,
    scale: 0.96,
    transition: { duration: 0.15, ease: 'easeIn' },
  },
}

export const overlayVariants: Variants = {
  initial: { opacity: 0 },
  animate: { opacity: 1, transition: { duration: 0.2 } },
  exit: { opacity: 0, transition: { duration: 0.15 } },
}

export const toastVariants: Variants = {
  initial: { opacity: 0, y: -12, scale: 0.95 },
  animate: {
    opacity: 1,
    y: 0,
    scale: 1,
    transition: { type: 'spring', stiffness: 400, damping: 28 },
  },
  exit: {
    opacity: 0,
    y: -8,
    scale: 0.96,
    transition: { duration: 0.15 },
  },
}
