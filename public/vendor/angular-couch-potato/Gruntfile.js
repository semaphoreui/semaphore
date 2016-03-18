module.exports = function(grunt) {

  grunt.loadNpmTasks('grunt-bower-task');
  grunt.renameTask('bower', 'bowerTask');
  grunt.loadNpmTasks('grunt-bower');
  grunt.renameTask('bower', 'gruntBower');
  grunt.loadNpmTasks('grunt-contrib-copy');
  // grunt.loadNpmTasks('grunt-contrib-jshint');
  grunt.loadNpmTasks('grunt-contrib-clean');
  grunt.loadNpmTasks('grunt-contrib-concat');
  // grunt.loadNpmTasks('grunt-contrib-uglify');
  grunt.loadNpmTasks('grunt-ngdocs');
  grunt.loadNpmTasks('grunt-gh-pages');

  // grunt.registerTask('bower', ['bowerTask', 'gruntBower']);
  grunt.registerTask('bower', ['bowerTask', 'gruntBower']);
  grunt.registerTask('default', ['build']);
  grunt.registerTask('build', ['clean', 'bower', 'concat']);
  grunt.registerTask('release', ['build','copy:samples', 'ngdocs']);
  grunt.registerTask('docs', ['ngdocs']);

  // Print a timestamp (useful for when watching)
  grunt.registerTask('timestamp', function() {
    grunt.log.subhead(Date());
  });

  // Project configuration.
  grunt.initConfig({
    dirs: {
      dist: 'dist',
      src: {
        js: ['src/**/couchPotato.js']
      }
    },
    'gh-pages': {
      options: {
        base: 'dist'
      },
      src: ['**']
    },
    pkg: grunt.file.readJSON('package.json'),
    banner:
    '/*! <%= pkg.title || pkg.name %> - v<%= pkg.version %> - <%= grunt.template.today("yyyy-mm-dd") %>\n' +
    '<%= pkg.homepage ? " * " + pkg.homepage + "\\n" : "" %>' +
    ' * Copyright (c) <%= grunt.template.today(\'yyyy\') %> <%= pkg.author.name %>;\n' +
    ' *    Uses software code originally found at https://github.com/szhanginrhythm/angular-require-lazyload\n' +
    ' * Licensed <%= _.pluck(pkg.licenses, "type").join(", ") %>\n */\n',
    ngdocs: {
      options: {
        dest: 'dist/docs',
        title: 'angular-couch-potato',
        startPage: '/guide',
        styles: ['docs/css/style.css'],
        navTemplate: 'docs/html/nav.html',
        html5Mode: false
      },
      guide: {
        src: ['docs/content/guide/**/*.ngdoc'],
        title: 'Guide'
      },
      api: {
        src: ['src/**/*.js', 'docs/content/api/**/*.ngdoc'],
        title: 'API Reference'
      }
    },
    clean: ['<%= dirs.dist %>/*'],
    gruntBower: {
      dev: {
        dest: '<%= dirs.dist %>/dependencies'
      }
    },
    bowerTask: {
      install: {
        options: {
          copy: false
        }
      }
    },
    copy: {
      samples: {
        expand: true,
        cwd: './',
        src: 'samples/**/*',
        dest: 'dist'
      }
    },
    concat: {
      dist: {
        options: {
          banner: '<%= banner %>',
          stripBanners: true
        },
        src:['<%= dirs.src.js %>'],
        dest:'<%= dirs.dist %>/<%= pkg.name %>.js'
      }
    }
  });
};
