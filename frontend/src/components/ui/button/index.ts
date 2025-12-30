import type { VariantProps } from "class-variance-authority"
import { cva } from "class-variance-authority"

export { default as Button } from "./Button.vue"

export const buttonVariants = cva(
  "inline-flex items-center justify-center gap-2 whitespace-nowrap rounded-xl text-sm font-semibold ring-offset-background transition-all duration-200 ease-out focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:pointer-events-none disabled:opacity-50 [&_svg]:pointer-events-none [&_svg]:size-4 [&_svg]:shrink-0 active:scale-95",
  {
    variants: {
      variant: {
        default: "bg-gradient-to-r from-primary via-primary/90 to-primary/80 text-primary-foreground hover:from-primary/95 hover:via-primary/85 hover:to-primary/75 shadow-xl hover:shadow-2xl hover:scale-105",
        destructive:
          "bg-gradient-to-r from-destructive via-destructive/90 to-destructive/80 text-destructive-foreground hover:from-destructive/95 hover:via-destructive/85 hover:to-destructive/75 shadow-xl hover:shadow-2xl hover:scale-105",
        outline:
          "border-2 border-primary/40 glass-button dark:glass-button hover:bg-primary/15 hover:border-primary/60 hover:text-primary hover:scale-105 hover:shadow-lg",
        secondary:
          "bg-gradient-to-r from-secondary via-secondary/90 to-secondary/80 text-secondary-foreground hover:from-secondary/95 hover:via-secondary/85 hover:to-secondary/75 shadow-lg hover:shadow-xl hover:scale-105",
        ghost: "hover:bg-primary/15 hover:text-primary hover:scale-105 backdrop-blur-sm",
        link: "text-primary underline-offset-4 hover:underline hover:text-primary/80 transition-colors",
      },
      size: {
        "default": "h-12 px-6 py-3 min-h-[44px]",
        "sm": "h-10 rounded-lg px-4 min-h-[40px]",
        "lg": "h-14 rounded-xl px-10 min-h-[56px] text-base",
        "icon": "h-12 w-12 min-h-[44px] min-w-[44px]",
        "icon-sm": "size-10 min-h-[40px] min-w-[40px]",
        "icon-lg": "size-14 min-h-[56px] min-w-[56px]",
      },
    },
    defaultVariants: {
      variant: "default",
      size: "default",
    },
  },
)

export type ButtonVariants = VariantProps<typeof buttonVariants>
