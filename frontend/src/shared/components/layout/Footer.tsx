import { Github, Mail } from 'lucide-react'

export const Footer = () => {
  const currentYear = new Date().getFullYear()

  return (
    <footer className="border-t border-white/10 bg-black/80 backdrop-blur-xl text-white">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-6 flex flex-col md:flex-row justify-between items-center gap-4">
        <div className="text-sm text-gray-300">
          © {currentYear} ICTU blockchain network. All rights reserved.
        </div>
        <div className="flex items-center gap-4 text-gray-200">
          <a
            href="https://github.com"
            target="_blank"
            rel="noopener noreferrer"
            className="p-2 rounded-full border border-white/15 bg-white/5 hover:bg-white/15 transition"
            aria-label="GitHub"
          >
            <Github className="h-5 w-5" />
          </a>
          <a
            href="mailto:support@ibn.vn"
            className="p-2 rounded-full border border-white/15 bg-white/5 hover:bg-white/15 transition"
            aria-label="Email"
          >
            <Mail className="h-5 w-5" />
          </a>
        </div>
      </div>
    </footer>
  )
}

