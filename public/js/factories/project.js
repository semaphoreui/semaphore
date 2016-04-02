app.factory('ProjectFactory', ['$http', function ($http) {
	var Project = function (project) {
		this.id = project.id;
		this.name = project.name;
	}

	Project.prototype.getURL = function () {
		return '/project/' + this.id;
	}

	return Project;
}]);