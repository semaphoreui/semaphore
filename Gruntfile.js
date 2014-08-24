module.exports = function(grunt) {
	require('load-grunt-tasks')(grunt);

	// Configuration
	grunt.initConfig({
		pkg: grunt.file.readJSON('package.json'),
		bump: {
			options: {
				files: ['package.json'],
				pushTo: 'origin',
				commitFiles: ['package.json']
			}
		},

		watch: {
			js: {
				files: ['public/js/*.js', 'public/js/**/*.js'],
				tasks: ['newer:copy:js']
			},
			styles: {
				files: ['public/css/{,**/}*.less'],
				tasks: ['newer:less:development']
			},
			livereload: {
				files: [
					'dist/{,**/}*.{css,js,png,jpg,jpeg,gif,webp,svg,html}'
				],
				options: {
					livereload: true
				}
			}
		},

		less: {
			development: {
				expand: true,
				cwd: 'public/css',
				dest: 'dist/css',
				src: '**/*',
				ext: '.css'
			},
			production: {
				expand: true,
				cwd: 'public/css',
				dest: 'dist/css',
				src: '**/*',
				ext: '.css',
				options: {
					cleancss: true
				}
			}
		},

		clean: {
			clean: {
				files: [{
					dot: true,
					src: [
						'dist'
					]
				}]
			}
		},

		// Copies remaining files to places other tasks can use
		copy: {
			js: {
				expand: true,
				cwd: 'public/js',
				dest: 'dist/js/',
				src: '{,**/}*.js'
			},
			img: {
				expand: true,
				cwd: 'public/img',
				dest: 'dist/img/',
				src: '{,**/}*.{png,jpg,jpeg,gif}'
			},
			vendor: {
				files: [{
					expand: true,
					cwd: 'public/vendor',
					dest: 'dist/vendor',
					src: ['**/*.js', '**/*.css', '**/*.png', '**/*.jpg', '**/*.jpeg', '**/*.woff', '**/*.ttf', '**/*.svg', '**/*.eot']
				}]
			},
			fonts: {
				expand: true,
				cwd: 'public/fonts',
				dest: 'dist/fonts/',
				src: '{,**/}*.{woff,ttf,svg,eot}'
			}
		},

		// Run some tasks in parallel to speed up the build process
		concurrent: {
			options: {
				limit: 6
			},
			server: [
				'copy:js',
				'copy:vendor',
				'copy:img',
				'copy:fonts'
			],
			watch: {
				tasks: [
					'nodemon:dev',
					'watch'
				],
				options: {
					logConcurrentOutput: true
				}
			}
		},

		uglify: {
			dist: {
				files: [{
					expand: true,
					cwd: 'public/js',
					src: ['**/*.js', '*.js'],
					dest: 'dist/js'
				}]
			}
		},

		nodemon: {
			dev: {
				script: 'lib/app.js',
				logConcurrentOutput: true,
				options: {
					cwd: __dirname,
					watch: ['lib/']
				}
			}
		}
	});
	
	grunt.registerTask('serve', [
		'clean:clean',
		'concurrent:server',
		'less:development',
		'concurrent:watch'
	]);

	grunt.registerTask('build', [
		'clean:clean',
		'concurrent:server',
		'less:production'
	]);

	grunt.registerTask('default', [
		'build'
	]);
};
