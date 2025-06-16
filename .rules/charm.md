### **Designing CLIs**

CLIs have the wonderful benefit of minimal user interface and the powerful ability to tap into the essence of the command line with pipelines. The lack of a GUI also means CLIs see faster development cycles, with the costs being opaqueness and learning curve for the user.

When building CLIs we ask questions like the following:
- Are these arguments and flags intuitive?
- What do helpful error messages look like? How can error messages help users figure out how to use the product?
- How can `--help` teach users how to get the most value out of this tool?
- Can this CLI become more powerful by leaning into pipelines?

We also strive to provide manpage entries and completions for popular shells.

---

## **The README**

The [__huh__](https://github.com/charmbracelet/huh) README is chock-full of examples and GIFs.

The README is critically important to the success of an open source product. It’s often a developer’s first point of contact with a project and the place where a developer will, in a matter of seconds, judge whether the project worthy of further consideration. With this in mind, put a lot of effort into README design, optimizing for strong first impressions.

Our strategy is to simply follow the age-old rule of advertising: showing the product. Good products, when presented correctly, will sell themselves, which is why we spend spend so much time on user experience and attention to detail.

With libraries, APIs, and packages we show the product with example code, typically placing some code right at the top of the README. We want to show the reader how intuitive, fun and powerful the API is and help them get started as quickly as possible.

Do you want fries with that?

When it makes sense, we also insert GIFs of the product right at the top of the README. While GIFs remain a technical nightmare in terms of a file format, they’re the most effective medium we’ve encountered for illustrating how software works in a concise manner. They’re short, silent videos that automatically autoplay and automatically loop with no user interaction, allowing us to paint a meaningful picture of the application in just a few seconds.

As mentioned earlier, we believe so much in the effectiveness of GIFs that we built [__vhs__](https://github.com/charmbracelet/vhs), a tool for scripting small, high quality GIFs of terminal sessions.

### **Quick Reference**

Beyond first impressions, we also tend to include a quick reference in our READMEs to support the docs. The quick reference gives developers more insight into the package, helps them get started, and highlights some of the common parts of an API. It’s an excellent bridge to the docs that can help the user make connections and gain insight into the API in ways the full, raw documentation cannot.

In go projects, the README is also browsable in the generated documentation (called GoDocs) so having the quick reference in the README is a win-win.

---

## **Examples, Examples, Examples**

An example in the [__Harmonica__](https://github.com/charmbracelet/harmonica) README

We can’t emphasize the effectiveness of examples enough.

We’re very strong believers in learning by example and we believe code examples are one of the best ways to learn about an API. They show developers how to use the API in a concrete way and help them hit the ground running so they can be more productive more quickly. Examples also serve as a cookbook, presenting the developers with solutions to common use cases, whetting their appetites for creative thinking with the product.

We commonly put a set of fully functional examples in a repository for users to reference both online and locally in their clones. In many cases those examples are what we ourselves used to help think about the product while we were building it.

---

## **Branding**

Good branding has the power to differentiate a product in the market with a mere glance. Our strategy is to appeal to developers on a personal level and create something that feels human, approachable, and memorable. We want the branding to stand apart from the common efforts we see from the vast majority of corporations and startups in the technology space. For that reason, we spend a lot of time looking beyond tech and instead drawing inspiration from things like [__video games__](https://animalcrossing.nintendo.com/), [__art__](https://www.are.na/christian-rocha/glitter-odyssey), [__beauty__](https://fentybeauty.com/), [__Family Mart__](https://www.family.co.jp/), [__Sanrio__](https://www.sanrio.com/) and so on.

The essence of our products’ branding is the name. Our goal is to disarm our readers, suggest how they should feel about our product, and make them smile. For these reasons an ideal Charm name is subversive, casual, and tongue in cheek. Sometimes a good name comes quickly. Other times it takes awhile. Because of this we start branding efforts early on in the product development cycle.