import { useEffect } from 'react'
import { createPortal } from 'react-dom'
import { ChevronLeft, ChevronRight, X } from 'lucide-react'

export default function ImagePreview({ images, index, onIndexChange, onClose }) {
  const hasImages = Array.isArray(images) && images.length > 0
  const currentIndex = hasImages ? Math.min(Math.max(index, 0), images.length - 1) : 0
  const currentImage = hasImages ? images[currentIndex] : ''

  useEffect(() => {
    if (!hasImages) return undefined

    const handleKeydown = (event) => {
      if (event.key === 'Escape') onClose()
      if (event.key === 'ArrowLeft') onIndexChange((currentIndex - 1 + images.length) % images.length)
      if (event.key === 'ArrowRight') onIndexChange((currentIndex + 1) % images.length)
    }

    window.addEventListener('keydown', handleKeydown)
    return () => window.removeEventListener('keydown', handleKeydown)
  }, [currentIndex, hasImages, images, onClose, onIndexChange])

  if (!hasImages) return null

  return createPortal(
    <div className="image-preview" role="dialog" aria-modal="true">
      <button className="preview-backdrop" type="button" aria-label="关闭图片预览" onClick={onClose} />
      <div className="preview-stage">
        <button className="preview-close icon-only" type="button" aria-label="关闭图片预览" onClick={onClose}>
          <X size={22} />
        </button>

        {images.length > 1 && (
          <button
            className="preview-arrow preview-prev icon-only"
            type="button"
            aria-label="上一张"
            onClick={() => onIndexChange((currentIndex - 1 + images.length) % images.length)}
          >
            <ChevronLeft size={28} />
          </button>
        )}

        <img src={currentImage} alt={`预览图片 ${currentIndex + 1}`} />

        {images.length > 1 && (
          <button
            className="preview-arrow preview-next icon-only"
            type="button"
            aria-label="下一张"
            onClick={() => onIndexChange((currentIndex + 1) % images.length)}
          >
            <ChevronRight size={28} />
          </button>
        )}

        <div className="preview-count">
          {currentIndex + 1} / {images.length}
        </div>
      </div>
    </div>,
    document.body
  )
}
