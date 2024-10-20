require 'mina/default'
require 'mina/git'

set :branch, 'main'
set :execution_mode, 'system'
set :os, ['darwin']
set :arch, ['amd64']

desc 'Release a new version'
task :release do
  run(:local) do
    version = ENV['version']

    if version.nil? || version.empty?
      abort "err: version not specified"
    end
    Gem::Version.new(version)

    # check if git now on main and 
    sh %{
      if [[ $(git branch --show-current) != "main" ]]; then
        echo "err: branch is not main"
        exit 1
      fi
    }

    # check if there's no unstaged changes
    sh %{
      if [[ `git status --porcelain` ]]; then
        echo "err unstaged changes found"
        exit 1
      fi
    }

    # build binaries
    fetch(:os).each do |os|
      fetch(:arch).each do |arch|
        dist_name = "rye-#{version}-#{os}-#{arch}"
        dist_dir = "./dist/#{dist_name}"

        # create directory
        sh %(mkdir -p #{dist_dir})

        # build
        sh %(BUILD=$(eval git rev-parse HEAD) && GOOS=#{os} GOARCH=#{arch} go build -o #{dist_dir}/rye -ldflags "-s -w -X github.com/larryzhao/rye.Build=$BUILD -X github.com/larryzhao/rye.Version=#{version}" ./cmd/cli/...)
        sh %(touch #{dist_dir}/LICENSE)
        sh %(touch #{dist_dir}/README.md)

        # create a tar file
        sh %(cd dist && tar czf #{dist_name}.tar.gz #{dist_name}) 

        # call gh to create a release
        sh %(gh release create v#{version})
        
        # upload artifacts to the release
        sh %(gh release upload v#{version} ./dist/#{dist_name}.tar.gz)
      end
    end
  end
end