---
name: mobile-design
description: Create distinctive, production-grade mobile interfaces with high design quality using Flutter. Use this skill when the user asks to build mobile apps, screens, widgets, animations, or UI components. Generates creative, polished Flutter code that avoids generic standard widget aesthetics.
license: Complete terms in LICENSE.txt
---

This skill guides the creation of distinctive, production-grade mobile interfaces using Flutter that avoid generic "standard widget" aesthetics. Implement real working code with exceptional attention to aesthetic details, animations, and creative choices.

The user provides mobile requirements: a screen, app flow, custom widget, or interface to build. They may include context about the purpose, audience, or technical constraints.

## Design Thinking

Before coding, understand the context and commit to a BOLD aesthetic direction:
- **Purpose**: What problem does this app solve? Who is holding the device?
- **Tone**: Pick an extreme: brutally minimal, maximalist chaos, retro-futuristic, organic/natural, luxury/refined, playful/toy-like, editorial/magazine, neumorphic, glassmorphic, industrial/utilitarian, etc. Use these for inspiration but design one that is true to the aesthetic direction.
- **Constraints**: Technical requirements (Flutter performance, platform specificities for iOS/Android, accessibility).
- **Differentiation**: What makes this interaction UNFORGETTABLE? What's the one gesture or transition someone will remember?

**CRITICAL**: Choose a clear conceptual direction and execute it with precision. Avoid the "default Material Design" look unless explicitly requested.

Then implement working Flutter code (Dart) that is:
- **Production-grade**: Proper state management structure (even if simple), error handling, and responsiveness.
- **Visually striking**: Custom painters, shaders, and non-standard layouts.
- **Cohesive**: Consistent design language across all widgets.
- **Meticulously refined**: Exact padding, typography, and timing.

## Mobile Aesthetics Guidelines

Focus on:
- **Typography**: Go beyond the default `Roboto` or `SF Pro`. Use `google_fonts` to introduce character. Pair distinctive display fonts with readable body text. Use varied weights and heights (`height` property in `TextStyle`) to create rhythm.
- **Color & Theme**: Commit to a cohesive `ThemeData`. Use `ColorScheme` semantics but push the boundaries. don't just use primary/secondary; use custom extensions if needed. Dominant colors with sharp accents outperform timid, safe palettes.
- **Motion & Interaction**: Flutter excels here. Use `AnimationController` for choreographed sequences. Implement `Hero` transitions for continuity. Use `AnimatedBuilder` for complex effects. Focus on the "feel" of touch—custom splashes, scale effects on press, and physics-based scrolling. One well-orchestrated screen transition creates more delight than scattered generic fades.
- **Spatial Composition**: Break the standard `Column`/`Row` grid. Use `Stack` and `Positioned` for layering. Use `Transform` for depth. distinct use of white space (or lack thereof).
- **Backgrounds & Visual Details**: Avoid flat solid backgrounds. Use `CustomPainter` for generative backgrounds, `ShaderMask` for gradients on text/icons, `BackdropFilter` for blur/glass effects, and noise textures. Add depth with careful shadows (`BoxShadow`) and layers.

NEVER use generic AI-generated aesthetics like standard unstyled implementations, cliched blue/purple gradients, or "default Flutter counter app" vibes.

Interpret creatively and make unexpected choices that feel genuinely designed for the context. No design should be the same. Vary between light and dark themes, different fonts, different aesthetics.

**IMPORTANT**: Match implementation complexity to the aesthetic vision.
- **Maximalist**: Heavy use of `CustomPainter`, `ShaderMask`, explicit animations, and layered `Stack` layouts.
- **Minimalist**: rigorous attention to `Padding`, `Align`, typography, and subtle implicit animations (`AnimatedContainer`, `AnimatedOpacity`).

Remember: Claude is capable of extraordinary creative work in Flutter. Don't hold back—show what can truly be created when thinking outside the box and committing fully to a distinctive vision.