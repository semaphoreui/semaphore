package projects

import (
	"database/sql"

	"github.com/ansible-semaphore/semaphore/database"
	"github.com/ansible-semaphore/semaphore/models"
	"github.com/ansible-semaphore/semaphore/util"
	"github.com/gin-gonic/gin"
	"github.com/masterminds/squirrel"
)

func RepositoryMiddleware(c *gin.Context) {
	project := c.MustGet("project").(models.Project)
	repositoryID, err := util.GetIntParam("repository_id", c)
	if err != nil {
		return
	}

	var repository models.Repository
	if err := database.Mysql.SelectOne(&repository, "select * from project__repository where project_id=? and id=?", project.ID, repositoryID); err != nil {
		if err == sql.ErrNoRows {
			c.AbortWithStatus(404)
			return
		}

		panic(err)
	}

	c.Set("repository", repository)
	c.Next()
}

func GetRepositories(c *gin.Context) {
	project := c.MustGet("project").(models.Project)
	var repos []models.Repository

	query, args, _ := squirrel.Select("*").
		From("project__repository").
		Where("project_id=?", project.ID).
		ToSql()

	if _, err := database.Mysql.Select(&repos, query, args...); err != nil {
		panic(err)
	}

	c.JSON(200, repos)
}

func AddRepository(c *gin.Context) {
	project := c.MustGet("project").(models.Project)

	var repository models.Repository
	if err := c.Bind(&repository); err != nil {
		return
	}

	if _, err := database.Mysql.Exec("insert into project__repository set project_id=?, git_url=?, ssh_key_id=?", project.ID, repository.GitUrl, repository.SshKeyID); err != nil {
		panic(err)
	}

	c.AbortWithStatus(204)
}

func UpdateRepository(c *gin.Context) {
	var repository models.Repository
	if err := c.Bind(&repository); err != nil {
		return
	}

	if _, err := database.Mysql.Exec("update project__repository set git_url=?, ssh_key_id=? where id=?", repository.GitUrl, repository.SshKeyID, c.MustGet("repository").(models.Repository).ID); err != nil {
		panic(err)
	}

	c.AbortWithStatus(204)
}

func RemoveRepository(c *gin.Context) {
	repository := c.MustGet("repository").(models.Repository)

	if _, err := database.Mysql.Exec("delete from project__repository where id=?", repository.ID); err != nil {
		panic(err)
	}

	c.AbortWithStatus(204)
}
