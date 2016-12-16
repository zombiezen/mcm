# mcm Documentation

Docs are built with [Jekyll](https://jekyllrb.com).

```bash
# Install RVM to ensure latest Ruby
gpg --keyserver hkp://keys.gnupg.net --recv-keys 409B6B1796C275462A1703113804BB82D39DC0E3
curl -sSL https://get.rvm.io | bash -s stable --ruby
. ~/.rvm/scripts/rvm

# Install local deps
gem install bundler
bundle install

# Local dev
bundle exec jekyll serve
```

You may also want to set `JEKYLL_GITHUB_TOKEN` for [github-metadata](https://github.com/jekyll/github-metadata).

## License

The docs use the [Hyde](http://hyde.getpoole.com/) theme, which is covered under an [MIT License](LICENSE.md).
The content is covered under the general project Apache 2.0 license.
