import { ReactNode } from 'react'
import { AnimatePresence, motion } from 'framer-motion'
import { dialogVariants, overlayVariants } from '../../lib/motionVariants'

interface AnimatedDialogProps {
  open: boolean
  onClose?: () => void
  children: ReactNode
  width?: string
  zIndex?: number
}

export default function AnimatedDialog({
  open,
  onClose,
  children,
  width = 'w-[460px]',
  zIndex = 50,
}: AnimatedDialogProps) {
  return (
    <AnimatePresence>
      {open && (
        <motion.div
          className="fixed inset-0 flex items-center justify-center"
          style={{ zIndex, backdropFilter: 'blur(4px)' }}
          variants={overlayVariants}
          initial="initial"
          animate="animate"
          exit="exit"
        >
          {/* Backdrop */}
          <div
            className="absolute inset-0"
            style={{ backgroundColor: 'rgba(0,0,0,0.65)' }}
            onClick={onClose}
          />

          {/* Dialog panel */}
          <motion.div
            className={`relative ${width} rounded-2xl p-6 shadow-dialog`}
            style={{
              background: 'var(--bg-elevated)',
              border: '1px solid var(--border-accent)',
              boxShadow: 'var(--shadow-dialog), var(--glow-accent-sm)',
            }}
            variants={dialogVariants}
            initial="initial"
            animate="animate"
            exit="exit"
            onClick={e => e.stopPropagation()}
          >
            {children}
          </motion.div>
        </motion.div>
      )}
    </AnimatePresence>
  )
}
